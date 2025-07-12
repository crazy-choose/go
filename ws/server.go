package ws

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var (
	ErrServerClosed  = errors.New("websocket server instance is closed")
	ErrNotFoundAlias = errors.New("websocket server not found alias")
)

type handleMessageFunc func(*Session, []byte)
type handleErrorFunc func(*Session, error)
type handleCloseFunc func(*Session, int, string) error
type handleSessionFunc func(*Session)
type filterFunc func(*Session) bool

type Server struct {
	Config                   *Config
	upGrader                 *websocket.Upgrader
	messageHandler           handleMessageFunc
	messageHandlerBinary     handleMessageFunc
	messageSentHandler       handleMessageFunc
	messageSentHandlerBinary handleMessageFunc
	errorHandler             handleErrorFunc
	closeHandler             handleCloseFunc
	connectHandler           handleSessionFunc
	disconnectHandler        handleSessionFunc
	pongHandler              handleSessionFunc
	bucket                   *bucket
}

func NewServer() *Server {
	return NewServerWithConfig(nil)
}

func NewServerWithConfig(cfg *Config) *Server {
	ug := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	b := new(bucket).init()
	go b.run()

	if cfg == nil {
		cfg = newConfig()
	}

	s := &Server{
		Config:                   cfg,
		upGrader:                 ug,
		messageHandler:           func(*Session, []byte) {},
		messageHandlerBinary:     func(*Session, []byte) {},
		messageSentHandler:       func(*Session, []byte) {},
		messageSentHandlerBinary: func(*Session, []byte) {},
		errorHandler:             func(*Session, error) {},
		closeHandler:             nil,
		connectHandler:           func(*Session) {},
		disconnectHandler:        func(*Session) {},
		pongHandler:              func(*Session) {},
		bucket:                   b,
	}

	return s
}

func (s *Server) HandleConnect(fn func(*Session)) {
	s.connectHandler = fn
}

func (s *Server) HandleDisconnect(fn func(*Session)) {
	s.disconnectHandler = fn
}

func (s *Server) HandlePong(fn func(*Session)) {
	s.pongHandler = fn
}

func (s *Server) HandleRequest(fn func(*Session, []byte)) {
	s.messageHandler = fn
}

func (s *Server) HandleRequestBinary(fn func(*Session, []byte)) {
	s.messageHandlerBinary = fn
}

func (s *Server) HandleSentMessage(fn func(*Session, []byte)) {
	s.messageSentHandler = fn
}

func (s *Server) HandleSentMessageBinary(fn func(*Session, []byte)) {
	s.messageSentHandlerBinary = fn
}

func (s *Server) HandleError(fn func(*Session, error)) {
	s.errorHandler = fn
}

func (s *Server) HandleClose(fn func(*Session, int, string) error) {
	if fn != nil {
		s.closeHandler = fn
	}
}

func (s *Server) HandleUpgrade(c *gin.Context) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	w := c.Writer
	r := c.Request
	conn, err := s.upGrader.Upgrade(w, r, w.Header())
	if err != nil {
		return err
	}

	session := &Session{
		UUID:       uuid.New(),
		GinContext: c,
		Alias:      "",
		tags:       make(map[string]interface{}),
		server:     s,
		conn:       conn,
		send:       make(chan *envelope, s.Config.MessageBufferSize),
		open:       true,
		rwMutex:    &sync.RWMutex{},
	}

	s.bucket.register <- session

	s.connectHandler(session)

	go session.writePump()

	session.readPump()

	session.Close()

	if session.closed() {
		s.bucket.unregister <- session
	}

	s.disconnectHandler(session)

	return nil
}

func (s *Server) GetAllSession() []Session {
	ret := make([]Session, 0)
	s.bucket.rwMux.RLock()
	for _, session := range s.bucket.sessions {
		ret = append(ret, Session{
			UUID:  session.UUID,
			Alias: session.Alias,
			tags:  session.tags,
		})
	}
	s.bucket.rwMux.RUnlock()
	return ret
}

func (s *Server) GetSession(uuidstr string) *Session {
	return s.bucket.get(uuidstr)
}

func (s *Server) GetTagsByAlias(alias string) []string {
	return s.bucket.getTagsByAlias(alias)
}

func (s *Server) GetAllAlias() []string {
	return s.bucket.allAlias()
}

func (s *Server) GetAllUUID() []string {
	return s.bucket.getAllUUID()
}

func (s *Server) Send(msg interface{}, alias string, tags ...string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	if alias == "" {
		return ErrNotFoundAlias
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &envelope{t: websocket.TextMessage, msg: b, filter: func(session *Session) bool {
		if session.Alias != alias {
			return false
		}

		if len(tags) < 1 {
			return true
		}

		return session.HaveTags(tags...)
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) SendBinary(msg []byte, alias string, tags ...string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	if alias == "" {
		return ErrNotFoundAlias
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg, filter: func(session *Session) bool {
		if session.Alias != alias {
			return false
		}

		if len(tags) < 1 {
			return true
		}

		return session.HaveTags(tags...)
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) Broadcast(msg interface{}) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &envelope{t: websocket.TextMessage, msg: b}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastByAlias(msg interface{}, alias string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	if alias == "" {
		return ErrNotFoundAlias
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &envelope{t: websocket.TextMessage, msg: b, filter: func(session *Session) bool {
		return session.Alias == alias
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastByTags(msg interface{}, tags ...string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &envelope{t: websocket.TextMessage, msg: b, filter: func(session *Session) bool {
		return session.HaveTags(tags...)
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastByTagsHaveAll(msg interface{}, tags ...string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	message := &envelope{t: websocket.TextMessage, msg: b, filter: func(session *Session) bool {
		return session.HaveAllTags(tags...)
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastBinary(msg []byte) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastBinaryByAlias(msg []byte, alias string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	if alias == "" {
		return ErrNotFoundAlias
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg, filter: func(session *Session) bool {
		return session.Alias == alias
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) BroadcastBinaryByTags(msg []byte, tags ...string) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg, filter: func(session *Session) bool {
		return session.HaveTags(tags...)
	}}
	s.bucket.broadcast <- message

	return nil
}

func (s *Server) SetTagsByAlias(alias string, tags ...string) {
	s.bucket.setTags(alias, tags...)
}

func (s *Server) DelTagsByAlias(alias string, tags ...string) {
	s.bucket.delTags(alias, tags...)
}

func (s *Server) Close() error {
	if s.bucket.closed() {
		return ErrServerClosed
	}
	s.bucket.exit <- &envelope{t: websocket.CloseMessage, msg: []byte{}}
	return nil
}

func (s *Server) CloseWithMsg(msg []byte) error {
	if s.bucket.closed() {
		return ErrServerClosed
	}
	s.bucket.exit <- &envelope{t: websocket.CloseMessage, msg: msg}
	return nil
}

func (s *Server) Len() int {
	return s.bucket.len()
}

func (s *Server) IsClosed() bool {
	return s.bucket.closed()
}
