package web

import (
	"go-uuid/uuid"
	"net/http"
	"github.com/quintans/toolkit/cache"
	"strings"
	"time"
)

const (
	GSESSION = "GSESSION"
)

type ISession interface {
	GetId() string
	Invalidate()
	IsInvalid() bool
	Get(key string) interface{}
	Delete(key string)
	Put(key string, value interface{})
}

func NewSession() *Session {
	this := new(Session)
	this.Init()
	return this
}

type Session struct {
	Id         string
	Attributes map[string]interface{}
	Invalid    bool // if true it will be invalidate (removal from Sessions) at the end of the request.
}

func (this *Session) Init() {
	this.Id = strings.Replace(uuid.New(), "-", "", -1)
	this.Attributes = make(map[string]interface{})
}

func (this *Session) GetId() string {
	return this.Id
}

func (this *Session) Invalidate() {
	this.Invalid = true
}

func (this *Session) IsInvalid() bool {
	return this.Invalid
}

func (this *Session) Get(key string) interface{} {
	v, ok := this.Attributes[key]
	if ok {
		return v
	}
	return nil
}

func (this *Session) Delete(key string) {
	delete(this.Attributes, key)
}

func (this *Session) Put(key string, value interface{}) {
	this.Attributes[key] = value
}

type Sessions struct {
	factory func() ISession
	cache   *cache.ExpirationCache
}

func NewSessions(timeout time.Duration, interval time.Duration, factory func() ISession) *Sessions {
	s := new(Sessions)
	if factory != nil {
		s.factory = factory
	} else {
		// default factory
		s.factory = func() ISession {
			return NewSession()
		}
	}
	s.cache = cache.NewExpirationCache(timeout, interval)
	return s
}

func (this *Sessions) createCookieAndSession(w http.ResponseWriter) ISession {
	s := this.CreateNewSession()

	c := &http.Cookie{Name: GSESSION, Value: s.GetId(), Path: "/"}
	http.SetCookie(w, c)

	return s
}

// gets the session identified by id.
// If not found and if 'create' equals to 'true' then creates a new session
func (this *Sessions) GetOrCreate(w http.ResponseWriter, r *http.Request, create bool) ISession {
	cookie, _ := r.Cookie(GSESSION)

	if cookie == nil {
		if create {
			session := this.createCookieAndSession(w)
			logger.Debugf("No session cookie. Creating new session with %s", session.GetId())
			return session
		}
	} else {
		// get web context
		s := this.cache.GetIfPresentAndTouch(cookie.Value)
		if s == nil {
			if create {
				session := this.createCookieAndSession(w)
				logger.Debugf("Invalid session cookie. Creating new session with %s", session.GetId())
				return session
			} else {
				logger.Debugf("No session was found for %s", cookie.Value)
			}
		} else {
			return s.(ISession)
		}
	}
	return nil
}

func (this *Sessions) Delete(r *http.Request) {
	cookie, _ := r.Cookie(GSESSION)
	if cookie != nil {
		this.cache.Delete(cookie.Value)
	}
}

func (this *Sessions) Invalidate(session ISession) {
	if session != nil {
		this.cache.Delete(session.GetId())
	}
}

func (this *Sessions) GetIfPresent(id string) ISession {
	s := this.cache.GetIfPresent(id)
	if s != nil {
		return s.(ISession)
	}
	return nil
}

func (this *Sessions) CreateNewSession() ISession {
	s := this.factory()
	this.cache.Put(s.GetId(), s)
	return s
}
