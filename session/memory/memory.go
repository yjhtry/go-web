package memory

import (
	"container/list"
	"sync"
	"time"

	"github.com/yjhtry/go-web/session"
)

var pder = &Provider{list: list.New()}

type SessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

func (s *SessionStore) Set(key, value interface{}) {
	pder.SessionUpdate(s.sid)
	s.value[key] = value
}

func (s *SessionStore) Get(key interface{}) interface{} {
	pder.SessionUpdate(s.sid)
	if v, ok := s.value[key]; ok {
		return v
	}
	return nil
}

func (s *SessionStore) Delete(key interface{}) {
	pder.SessionUpdate(s.sid)
	delete(s.value, key)
}

func (s *SessionStore) SessionID() string {
	return s.sid
}

type Provider struct {
	lock     sync.Mutex
	list     *list.List
	sessions map[string]*list.Element
}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	v := make(map[interface{}]interface{}, 0)

	session := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := pder.list.PushBack(session)

	pder.sessions[sid] = element

	return session, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := pder.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		session, err := pder.SessionInit(sid)
		return session, err
	}
}

func (pder *Provider) SessionDestroy(sid string) error {
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
	}

	return nil
}

func (pder *Provider) SessionGC(maxLifeTime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}

		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
	}

	return nil
}

func init() {
	pder.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", pder)
}
