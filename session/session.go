package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var provides = make(map[string]Provider)

type Manager struct {
	cookieName  string     // private cookie name
	lock        sync.Mutex // protects session
	provider    Provider
	maxLifeTime int64
}

func (manager *Manager) SessionId() string {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	cookie, err := r.Cookie(manager.cookieName)

	if err != nil || cookie.Value == "" {
		sid := manager.SessionId()
		session, _ = manager.provider.SessionInit(sid)

		cookie := http.Cookie{
			Name:     manager.cookieName,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(manager.maxLifeTime),
		}

		http.SetCookie(w, &cookie)

	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRead(sid)
	}

	return
}

func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)

	if err != nil || cookie.Value == "" {
		return
	} else {

		manager.lock.Lock()
		defer manager.lock.Unlock()

		manager.provider.SessionDestroy(cookie.Value)

		expiration := time.Now()
		cookie := http.Cookie{
			Name:     manager.cookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1,
		}
		http.SetCookie(w, &cookie)
	}
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	time.AfterFunc(time.Duration(manager.maxLifeTime)*time.Second, func() {
		manager.provider.SessionGC(manager.maxLifeTime)
	})
}

func NewSessionManager(provideName, cookieName string, maxLifeTime int64) (*Manager, error) {
	provider, ok := provides[provideName]

	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	return &Manager{
		provider:    provider,
		cookieName:  cookieName,
		maxLifeTime: maxLifeTime}, nil
}

func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provider is nil")
	}

	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}

	provides[name] = provider
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

type Session interface {
	Get(key interface{}) interface{}
	Set(key, value interface{})
	Delete(key interface{})
	SessionID() string
}
