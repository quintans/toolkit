package web

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/quintans/toolkit/cache"

	"net/http"
	"strings"
	"time"
)

const (
	COOKIE_NAME = "GSESSION"
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

type SessionsConfig struct {
	Timeout    time.Duration
	Interval   time.Duration
	Factory    func() ISession
	CookieName string
}

type Sessions struct {
	config SessionsConfig
	cache  *cache.ExpirationCache
}

func NewSessions(config SessionsConfig) *Sessions {
	s := new(Sessions)
	if int64(config.Timeout) == 0 {
		config.Timeout = 20 * time.Minute
	}
	if int64(config.Interval) == 0 {
		config.Interval = time.Minute
	}
	if config.Factory == nil {
		// default factory
		config.Factory = func() ISession {
			return NewSession()
		}
	}
	if config.CookieName == "" {
		config.CookieName = COOKIE_NAME
	}
	s.config = config
	s.cache = cache.NewExpirationCache(config.Timeout, config.Interval)
	return s
}

func (this *Sessions) createCookieAndSession(w http.ResponseWriter) ISession {
	s := this.CreateNewSession()

	c := &http.Cookie{Name: this.config.CookieName, Value: s.GetId(), Path: "/"}
	http.SetCookie(w, c)

	return s
}

// gets the session identified by id.
// If not found and if 'create' equals to 'true' then creates a new session
func (this *Sessions) GetOrCreate(w http.ResponseWriter, r *http.Request, create bool) ISession {
	cookie, _ := r.Cookie(this.config.CookieName)

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
	cookie, _ := r.Cookie(this.config.CookieName)
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
	s := this.config.Factory()
	this.cache.Put(s.GetId(), s)
	return s
}
