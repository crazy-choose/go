package ws

import (
	"encoding/json"
	"errors"
	"github.com/crazy-choose/helper/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Session struct {
	uuid.UUID
	GinContext *gin.Context
	Alias      string
	tags       map[string]interface{}
	server     *Server
	conn       *websocket.Conn
	send       chan *envelope
	once       sync.Once
	open       bool
	rwMutex    *sync.RWMutex
}

func (s *Session) AddTags(tags ...string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	for _, tag := range tags {
		s.tags[tag] = struct{}{}
	}
}

func (s *Session) DelTags(tags ...string) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	for _, tag := range tags {
		delete(s.tags, tag)
	}
}

func (s *Session) HaveTags(tags ...string) bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	for _, tag := range tags {
		if _, ok := s.tags[tag]; ok {
			return true
		}
	}
	return false
}

func (s *Session) HaveAllTags(tags ...string) bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	if len(tags) == 0 {
		return false
	}

	for _, tag := range tags {
		if _, ok := s.tags[tag]; !ok {
			return false
		}
	}
	return true
}

func (s *Session) Response(message string) {
	s.writeMessage(&envelope{
		t:      websocket.TextMessage,
		msg:    []byte(message),
		filter: nil,
	})
}

func (s *Session) ResponseBinary(message []byte) {
	s.writeMessage(&envelope{
		t:      websocket.BinaryMessage,
		msg:    message,
		filter: nil,
	})
}

func (s *Session) ResponseJson(v interface{}) {
	message, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.writeMessage(&envelope{
		t:      websocket.BinaryMessage,
		msg:    message,
		filter: nil,
	})
}

func (s *Session) ResponseJsonString(v interface{}) {
	message, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.writeMessage(&envelope{
		t:      websocket.TextMessage,
		msg:    message,
		filter: nil,
	})
}

func (s *Session) writeMessage(message *envelope) {
	if s.closed() {
		s.server.errorHandler(s, errors.New("tried to write to closed a session"))
		return
	}

	select {
	case s.send <- message:
		log.Debug("writeMessage(ch<-msg) uuid:%s alias:%s mt:%d msg:%s", s.UUID.String(), s.Alias, message.t, string(message.msg))
	default:
		log.Debug("writeMessage(ch<-msg) uuid:%s alias:%s mt:%d msg:%s err:send channel is full", s.UUID.String(), s.Alias, message.t, string(message.msg))
	}
}

func (s *Session) writeRaw(message *envelope) error {
	if s.closed() {
		return errors.New("tried to write to a closed session")
	}

	e := s.conn.SetWriteDeadline(time.Now().Add(s.server.Config.WriteWait))
	if e != nil {
		return e
	}
	return s.conn.WriteMessage(message.t, message.msg)
}

func (s *Session) closed() bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return !s.open
}

func (s *Session) Close() {
	if !s.closed() {
		s.rwMutex.Lock()
		s.open = false
		_ = s.conn.Close()
		close(s.send)
		s.rwMutex.Unlock()
	}
}

func (s *Session) ping() {
	if err := s.conn.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(time.Second)); err != nil {
		log.Debug("ping(ctl) uuid:%s alias:%s err:%s", s.UUID.String(), s.Alias, err.Error())
	} else {
		log.Debug("ping(ctl) uuid:%s alias:%s", s.UUID.String(), s.Alias)
	}
	if err := s.writeRaw(&envelope{t: websocket.TextMessage, msg: []byte("ping")}); err != nil {
		log.Debug("ping(txt) uuid:%s alias:%s err:%s", s.UUID.String(), s.Alias, err.Error())
	} else {
		log.Debug("ping(txt) uuid:%s alias:%s", s.UUID.String(), s.Alias)
	}
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.server.Config.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case msg, ok := <-s.send:
			if !ok {
				break loop
			}

			err := s.writeRaw(msg)
			if err != nil {
				log.Error("writeMessage(msg<-ch) uuid:%s alias:%s mt:%d msg:%s err:%s", s.UUID.String(), s.Alias, msg.t, string(msg.msg), err.Error())
			} else {
				log.Debug("writeMessage(msg<-ch) uuid:%s alias:%s mt:%d msg:%s", s.UUID.String(), s.Alias, msg.t, string(msg.msg))
			}

			if err != nil {
				s.server.errorHandler(s, err)
				break loop
			}

			if msg.t == websocket.CloseMessage {
				break loop
			}

			if msg.t == websocket.TextMessage {
				s.server.messageSentHandler(s, msg.msg)
			}

			if msg.t == websocket.BinaryMessage {
				s.server.messageSentHandlerBinary(s, msg.msg)
			}
		case <-ticker.C:
			s.ping()
		}
	}
}

func (s *Session) readPump() {
	s.conn.SetReadLimit(s.server.Config.MaxMessageSize)

	_ = s.conn.SetReadDeadline(time.Now().Add(s.server.Config.PongWait))
	s.conn.SetPongHandler(func(appData string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.server.Config.PongWait))
		s.server.pongHandler(s)
		return nil
	})

	if s.server.closeHandler != nil {
		s.conn.SetCloseHandler(func(code int, text string) error {
			return s.server.closeHandler(s, code, text)
		})
	}

	for {
		t, message, err := s.conn.ReadMessage()

		if err != nil {
			s.server.errorHandler(s, err)
			break
		}

		if t == websocket.TextMessage {
			s.server.messageHandler(s, message)
		}

		if t == websocket.BinaryMessage {
			s.server.messageHandlerBinary(s, message)
		}
	}
}
