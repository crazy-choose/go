package ws

import (
	"github.com/google/uuid"
	"sync"
)

type bucket struct {
	open       bool
	register   chan *Session
	unregister chan *Session
	sessions   map[uuid.UUID]*Session
	broadcast  chan *envelope
	exit       chan *envelope
	rwMux      sync.RWMutex
}

func (b *bucket) init() *bucket {
	b.open = true
	b.register = make(chan *Session)
	b.unregister = make(chan *Session)
	b.sessions = make(map[uuid.UUID]*Session)
	b.broadcast = make(chan *envelope)
	b.exit = make(chan *envelope)
	b.rwMux = sync.RWMutex{}

	return b
}

func (b *bucket) get(key string) *Session {
	uuidObj, err := uuid.Parse(key)
	if err != nil {
		return nil
	}

	b.rwMux.RLock()
	s, ok := b.sessions[uuidObj]
	b.rwMux.RUnlock()

	if ok {
		return s
	} else {
		return nil
	}
}

func (b *bucket) allAlias() []string {
	var ret []string
	b.rwMux.RLock()
	for _, session := range b.sessions {
		ret = append(ret, session.Alias)
	}
	b.rwMux.RUnlock()
	return ret
}

func (b *bucket) getTagsByAlias(alias string) []string {
	var ret []string
	b.rwMux.RLock()
	for _, session := range b.sessions {
		if session.Alias == alias {
			for tag := range session.tags {
				ret = append(ret, tag)
			}
			break
		}
	}
	b.rwMux.RUnlock()
	return ret
}

func (b *bucket) getAllUUID() []string {
	var ret []string
	b.rwMux.RLock()
	for uid := range b.sessions {
		ret = append(ret, uid.String())
	}
	b.rwMux.RUnlock()
	return ret
}

func (b *bucket) run() {
loop:
	for {
		select {
		case s := <-b.register:
			b.doRegister(s)
		case s := <-b.unregister:
			b.doUnregister(s)
		case m := <-b.broadcast:
			b.doBroadcast(m)
		case m := <-b.exit:
			b.doExit(m)
			break loop
		}
	}
}

func (b *bucket) closed() bool {
	b.rwMux.RLock()
	defer b.rwMux.RUnlock()
	return !b.open
}

func (b *bucket) len() int {
	b.rwMux.RLock()
	defer b.rwMux.RUnlock()

	return len(b.sessions)
}

func (b *bucket) doRegister(session *Session) {
	b.rwMux.Lock()
	b.sessions[session.UUID] = session
	b.rwMux.Unlock()
}

func (b *bucket) doUnregister(session *Session) {
	if _, ok := b.sessions[session.UUID]; ok {
		b.rwMux.Lock()
		delete(b.sessions, session.UUID)
		b.rwMux.Unlock()
	}
}

func (b *bucket) setTags(alias string, tags ...string) {
	b.rwMux.Lock()
	for _, session := range b.sessions {
		if session.Alias == alias {
			session.AddTags(tags...)
		}
	}
	b.rwMux.Unlock()
}

func (b *bucket) delTags(alias string, tags ...string) {
	b.rwMux.Lock()
	for _, session := range b.sessions {
		if session.Alias == alias {
			session.DelTags(tags...)
		}
	}
	b.rwMux.Unlock()
}

func (b *bucket) doBroadcast(msg *envelope) {
	b.rwMux.RLock()
	for _, session := range b.sessions {
		if msg.filter != nil {
			if msg.filter(session) {
				session.writeMessage(msg)
			}
		} else {
			session.writeMessage(msg)
		}
	}
	b.rwMux.RUnlock()
}

func (b *bucket) doExit(msg *envelope) {
	b.rwMux.Lock()
	for uid, session := range b.sessions {
		session.writeMessage(msg)
		delete(b.sessions, uid)
		session.Close()
	}
	b.open = false
	b.rwMux.Unlock()
}
