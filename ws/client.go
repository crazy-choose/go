package ws

import (
	"encoding/json"
	"errors"
	"github.com/crazy-choose/helper/log"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	ConnectionNew      = "connection_new"
	ConnectionOpened   = "connection_opened"
	ConnectionClosed   = "connection_closed"
	ConnectionTryAgain = "connection_try_again"
)

const (
	TimerIntervalSecond     = 300 * time.Second
	ReconnectIntervalSecond = 5 * time.Second
	HeartbeatIntervalSecond = 60 * time.Second
)

type ConnectedHandler func(c *Connect)
type Heartbeat func(c *Connect)
type TextMessageHandler func(c *Connect, text string) (interface{}, error)
type BinaryMessageHandler func(c *Connect, binary []byte) (interface{}, error)
type PingMessageHandler func(c *Connect)
type ResponseHandler func(c *Connect, response interface{})

type Connect struct {
	path                 string
	header               http.Header
	conn                 *websocket.Conn
	connectedHandler     ConnectedHandler
	heartbeat            Heartbeat
	heartbeatInterval    time.Duration
	textMessageHandler   TextMessageHandler
	binaryMessageHandler BinaryMessageHandler
	pingMessageHandler   PingMessageHandler
	responseHandler      ResponseHandler
	stopTickerCh         chan bool
	reconnectCh          chan bool
	lastReceivedTime     time.Time
	sendMutex            *sync.Mutex
	option               Option
	status               string
}

type Option struct {
	Name                string // only show log
	ReconnectWaitSecond time.Duration
}

func NewClient(path string, header http.Header, option Option) *Connect {
	return &Connect{
		path:                 path,
		header:               header,
		stopTickerCh:         make(chan bool),
		reconnectCh:          make(chan bool),
		sendMutex:            &sync.Mutex{},
		connectedHandler:     func(c *Connect) {},
		heartbeat:            func(c *Connect) {},
		heartbeatInterval:    HeartbeatIntervalSecond,
		textMessageHandler:   func(c *Connect, text string) (interface{}, error) { return nil, nil },
		binaryMessageHandler: func(c *Connect, binary []byte) (interface{}, error) { return nil, nil },
		pingMessageHandler:   func(c *Connect) {},
		responseHandler:      func(c *Connect, response interface{}) {},
		option:               option,
		status:               ConnectionNew,
	}
}

func (c *Connect) ConnectAwaitSuccess() {
	c.connectInfiniteRetry()
	go c.tickerLoop()
}

func (c *Connect) Connect() error {
	if err := c.connect(); err != nil {
		return err
	}
	go c.tickerLoop()
	return nil
}

func (c *Connect) Status() string {
	return c.status
}

func (c *Connect) Heartbeat(fn func(c *Connect), interval time.Duration) {
	c.heartbeat = fn
	c.heartbeatInterval = interval
}

func (c *Connect) SetHandler(connHandler ConnectedHandler, textHandler TextMessageHandler, binaryHandler BinaryMessageHandler, respHandler ResponseHandler) {
	if connHandler != nil {
		c.connectedHandler = connHandler
	}

	if textHandler != nil {
		c.textMessageHandler = textHandler
	}

	if binaryHandler != nil {
		c.binaryMessageHandler = binaryHandler
	}

	if respHandler != nil {
		c.responseHandler = respHandler
	}
}

func (c *Connect) SetPingHandler(pingHandler PingMessageHandler) {
	if pingHandler != nil {
		c.pingMessageHandler = pingHandler
	}
}

func (c *Connect) SendString(data []byte) error {
	if c.conn == nil {
		return errors.New("no connection available")
	}

	c.sendMutex.Lock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	c.sendMutex.Unlock()
	return err
}

func (c *Connect) SendBinary(data []byte) error {
	if c.conn == nil {
		return errors.New("no connection available")
	}

	c.sendMutex.Lock()
	err := c.conn.WriteMessage(websocket.BinaryMessage, data)
	c.sendMutex.Unlock()
	return err
}

func (c *Connect) SendJson(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendBinary(b)
}

func (c *Connect) SendJsonString(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendString(b)
}

func (c *Connect) Close() {
	c.stopTicker()
	c.disconnect()
	c.status = ConnectionClosed
}

func (c *Connect) connectInfiniteRetry() {
	var (
		rc  int
		err error
	)
	for {
		err = c.connect()
		if err == nil {
			break
		}
		time.Sleep(ReconnectIntervalSecond)
		rc++
		c.status = ConnectionTryAgain
		log.Debug("[WebSocket Client %s] connect retry %d", c.option.Name, rc)
	}
}

func (c *Connect) connect() error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.path, c.header)
	if err != nil {
		log.Debug("[WebSocket Client %s] connect failed err:%s", c.option.Name, err.Error())
		return err
	}
	log.Info("[WebSocket Client %s] connect success", c.option.Name)
	c.lastReceivedTime = time.Now()

	go c.readLoop()

	c.connectedHandler(c)

	c.status = ConnectionOpened
	return nil
}

func (c *Connect) disconnect() {
	if c.conn == nil {
		return
	}

	err := c.conn.Close()
	if err != nil {
		log.Error("[WebSocket Client %s] disconnect error: %s", c.option.Name, err)
		return
	}

	log.Info("[WebSocket Client %s] disconnected", c.option.Name)
}

func (c *Connect) stopTicker() {
	c.stopTickerCh <- true
}

func (c *Connect) tickerLoop() {
	log.Info("[WebSocket Client %s] ticker loop started", c.option.Name)
	ticker := time.NewTicker(TimerIntervalSecond)
	defer ticker.Stop()

	hbtk := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hbtk.C:
			c.heartbeat(c)
		// Receive tick from tickChannel
		case <-ticker.C:
			if c.option.ReconnectWaitSecond.Seconds() > 0 {
				elapsedSecond := time.Now().Sub(c.lastReceivedTime).Seconds()
				log.Debug("[WebSocket Client %s] received data %f sec ago", c.option.Name, elapsedSecond)

				if elapsedSecond > c.option.ReconnectWaitSecond.Seconds() {
					log.Debug("[WebSocket Client %s] longtime no receive data, do reconnect...", c.option.Name)
					c.disconnect()
					c.connectInfiniteRetry()
				}
			}
		case <-c.reconnectCh:
			log.Debug("[WebSocket Client %s] connection loss, do reconnect...", c.option.Name)
			c.disconnect()
			time.Sleep(c.option.ReconnectWaitSecond)
			c.connectInfiniteRetry()
		case <-c.stopTickerCh:
			log.Debug("[WebSocket Client %s] ticker loop stopped", c.option.Name)
			return
		}
	}
}

func (c *Connect) readLoop() {
	log.Info("[WebSocket Client %s] read loop started", c.option.Name)
	for {
		if c.conn == nil {
			log.Error("[WebSocket Client %s] read error: no connection object is null", c.option.Name)
			time.Sleep(time.Second)
			continue
		}

		msgType, buf, err := c.conn.ReadMessage()
		if err != nil {
			log.Error("[WebSocket Client %s] read error: %s", c.option.Name, err)
			c.reconnectCh <- true
			break
		}

		c.lastReceivedTime = time.Now()

		var result interface{}
		if msgType == websocket.BinaryMessage {
			result, err = c.binaryMessageHandler(c, buf)
		} else if msgType == websocket.TextMessage {
			result, err = c.textMessageHandler(c, string(buf))
		} else if msgType == websocket.PingMessage {
			c.pingMessageHandler(c)
			continue
		}

		if err != nil {
			log.Error("[WebSocket Client %s] handle message error: %s", c.option.Name, err)
			continue
		}

		c.responseHandler(c, result)
	}

	log.Info("[WebSocket Client %s] read loop stopped", c.option.Name)
}
