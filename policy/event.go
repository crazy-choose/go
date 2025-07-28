package policy

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/crazy-choose/go/log"
	"os"
	"sync"
	"time"
)

// 事件类型
const (
	EventOpenAllow = iota
	EventCloseProhibit
	EventForceClose
)

// MaxEventQueueSize 最大事件队列大小
const MaxEventQueueSize = 4092

// ScheduledEvent 定时事件
type ScheduledEvent struct {
	timestamp time.Time
	key       string
	eventType int
	index     int
}

// EventQueue 事件队列（最小堆）
type EventQueue []*ScheduledEvent

func (eq *EventQueue) Len() int           { return len(*eq) }
func (eq *EventQueue) Less(i, j int) bool { return (*eq)[i].timestamp.Before((*eq)[j].timestamp) }
func (eq *EventQueue) Swap(i, j int) {
	(*eq)[i], (*eq)[j] = (*eq)[j], (*eq)[i]
	(*eq)[i].index = i
	(*eq)[j].index = j
}

func (eq *EventQueue) Push(x interface{}) {
	n := len(*eq)
	event := x.(*ScheduledEvent)
	event.index = n
	*eq = append(*eq, event)
}

func (eq *EventQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	event := old[n-1]
	old[n-1] = nil
	event.index = -1
	*eq = old[0 : n-1]
	return event
}

// TimeEvent 交易时间管理器
type TimeEvent struct {
	mu           sync.RWMutex
	callbackMu   sync.RWMutex
	eventQueue   EventQueue
	eventChan    chan *ScheduledEvent
	stopChan     chan struct{}
	running      bool
	callbacks    map[string]map[int]func(int) // 支持多个回调
	isIO         bool
	maxQueueSize int
}

// NewTimeEvent 创建新的交易时间管理器
func NewTimeEvent(isIO bool) *TimeEvent {
	return &TimeEvent{
		eventQueue:   make(EventQueue, 0),
		eventChan:    make(chan *ScheduledEvent, 1024), // 缓冲通道，处理高并发
		stopChan:     make(chan struct{}),
		running:      false,
		callbacks:    make(map[string]map[int]func(int)),
		isIO:         isIO,
		maxQueueSize: MaxEventQueueSize,
	}
}

// Start 启动管理器
func (ttm *TimeEvent) Start() {
	ttm.mu.Lock()
	defer ttm.mu.Unlock()

	if ttm.running {
		return
	}

	ttm.running = true
	go ttm.eventLoop()
}

// Stop 停止管理器
func (ttm *TimeEvent) Stop() {
	ttm.mu.Lock()
	defer ttm.mu.Unlock()

	if !ttm.running {
		return
	}

	close(ttm.stopChan)
	ttm.running = false

	// 清理剩余事件并触发回调
	for len(ttm.eventQueue) > 0 {
		event := heap.Pop(&ttm.eventQueue).(*ScheduledEvent)
		ttm.fireCallbacks(event.key, event.eventType)
	}
}

// AddEvent 添加事件并注册回调
func (ttm *TimeEvent) AddEvent(key string, eventType int, eventTime time.Time, dur time.Duration, cb func(int)) error {
	ttm.mu.Lock()
	defer ttm.mu.Unlock()

	// 参数校验
	if key == "" || eventType < EventOpenAllow || eventType > EventForceClose {
		return fmt.Errorf("invalid key or event type: key=%s, eventType=%d", key, eventType)
	}
	if dur < 0 {
		return fmt.Errorf("duration cannot be negative: %v", dur)
	}

	// 注册回调函数（支持多个回调）
	if cb != nil {
		ttm.callbackMu.Lock()
		m, ok := ttm.callbacks[key]
		if !ok {
			m = make(map[int]func(int))
		}
		m[eventType] = cb
		ttm.callbacks[key] = m
		ttm.callbackMu.Unlock()
	}

	// 检查队列大小
	if len(ttm.eventQueue) >= ttm.maxQueueSize {
		return fmt.Errorf("event queue size limit reached: %d", ttm.maxQueueSize)
	}

	// 去重逻辑
	for _, event := range ttm.eventQueue {
		if event.key == key && event.eventType == eventType && event.timestamp.Equal(eventTime) {
			return nil // 事件已存在，忽略
		}
	}

	now := time.Now()
	// 计算触发时间
	var triggerTime time.Time
	switch eventType {
	case EventOpenAllow:
		triggerTime = eventTime.Add(dur)
	case EventCloseProhibit, EventForceClose:
		triggerTime = eventTime.Add(-dur)
	default:
		return fmt.Errorf("unknown event type: %d", eventType)
	}

	// 添加事件到队列（仅当事件未过期）
	if now.Before(triggerTime) {
		event := &ScheduledEvent{
			timestamp: triggerTime,
			key:       key,
			eventType: eventType,
		}
		heap.Push(&ttm.eventQueue, event)
		select {
		case ttm.eventChan <- event:
		default:
			// 通道满时直接处理
			ttm.fireCallbacks(key, eventType)
		}
	} else {
		// 过期事件直接触发回调
		time.AfterFunc(5*time.Second, func() {
			ttm.fireCallbacks(key, eventType)
		})
	}

	return nil
}

// CleanExpiredEvents 清理过期事件
func (ttm *TimeEvent) CleanExpiredEvents() {
	ttm.mu.Lock()
	defer ttm.mu.Unlock()

	now := time.Now()
	for len(ttm.eventQueue) > 0 && ttm.eventQueue[0].timestamp.Before(now) {
		event := heap.Pop(&ttm.eventQueue).(*ScheduledEvent)
		ttm.fireCallbacks(event.key, event.eventType)
	}
}

// 触发回调函数
func (ttm *TimeEvent) fireCallbacks(key string, eventType int) {
	ttm.callbackMu.RLock()
	defer ttm.callbackMu.RUnlock()

	cbs, exists := ttm.callbacks[key]
	if !exists {
		return
	}

	cb, ok := cbs[eventType]
	if !ok {
		return
	}
	// 异步触发回调
	cb(eventType)
	log.Debug("[TimeEvent],key:%s, et:%v cb trigger")

}

// 事件处理循环
func (ttm *TimeEvent) eventLoop() {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		ttm.mu.Lock()

		// 批量处理同一时间点的事件
		now := time.Now()
		var batch []*ScheduledEvent
		for len(ttm.eventQueue) > 0 && ttm.eventQueue[0].timestamp.Before(now) {
			event := heap.Pop(&ttm.eventQueue).(*ScheduledEvent)
			batch = append(batch, event)
		}

		// 设置定时器到下一个事件
		var nextEvent *ScheduledEvent
		if len(ttm.eventQueue) > 0 {
			nextEvent = ttm.eventQueue[0]
			timer.Reset(nextEvent.timestamp.Sub(now))
		} else {
			timer.Stop()
		}
		ttm.mu.Unlock()

		// 触发批量事件的回调
		for _, event := range batch {
			ttm.fireCallbacks(event.key, event.eventType)
		}

		select {
		case <-ttm.stopChan:
			return
		case event := <-ttm.eventChan:
			// 处理通道中的新事件
			ttm.mu.Lock()
			if ttm.running && now.Before(event.timestamp) {
				heap.Push(&ttm.eventQueue, event)
				if len(ttm.eventQueue) == 1 || event.timestamp.Before(ttm.eventQueue[0].timestamp) {
					timer.Reset(event.timestamp.Sub(time.Now()))
				}
			}
			ttm.mu.Unlock()
		case <-timer.C:
			// 定时器触发后继续循环
		}
	}
}

// SaveEvents 保存事件队列到文件
func (ttm *TimeEvent) SaveEvents(filename string) error {
	if !ttm.isIO {
		return nil
	}

	ttm.mu.RLock()
	defer ttm.mu.RUnlock()

	data, err := json.Marshal(ttm.eventQueue)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	// 写入临时文件后重命名，确保原子性
	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write events to temp file: %w", err)
	}
	if err := os.Rename(tmpFile, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// LoadEvents 从文件恢复事件队列
func (ttm *TimeEvent) LoadEvents(filename string) error {
	if !ttm.isIO {
		return nil
	}

	ttm.mu.Lock()
	defer ttm.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read events file: %w", err)
	}

	var events []*ScheduledEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}

	// 恢复事件队列，忽略过期事件
	now := time.Now()
	for _, event := range events {
		if event.timestamp.After(now) {
			heap.Push(&ttm.eventQueue, event)
		} else {
			ttm.fireCallbacks(event.key, event.eventType)
		}
	}

	return nil
}
