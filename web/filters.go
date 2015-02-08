package web

import (
	tk "github.com/quintans/toolkit"
	"net/http"
	"strings"
)

func NewHttpFail(status int, code string, message string) *HttpFail {
	this := new(HttpFail)
	this.Status = status
	this.Fail = new(tk.Fail)
	this.Code = code
	this.Fail.Message = message
	return this
}

type HttpFail struct {
	*tk.Fail

	Status int
}

type IContext interface {
	Proceed() error
	GetResponse() http.ResponseWriter
	GetRequest() *http.Request
	GetSession() ISession
	SetSession(ISession)
	GetPrincipal() interface{}
	SetPrincipal(interface{})
	GetAttribute(string) interface{}
	SetAttribute(string, interface{})
	GetCurrentFilter() *Filter
	SetCurrentFilter(*Filter)
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	this := new(Context)
	this.Init(this, w, r)
	return this
}

var _ IContext = &Context{}

type Context struct {
	Response      http.ResponseWriter
	Request       *http.Request
	Session       ISession
	Principal     interface{}            // user data
	Attributes    map[string]interface{} // attributes only valid in this request
	CurrentFilter *Filter
	overrider     IContext
}

func (this *Context) Init(ctx IContext, w http.ResponseWriter, r *http.Request) {
	this.overrider = ctx
	this.Response = w
	this.Request = r
	this.Attributes = make(map[string]interface{})
}

func (this *Context) Proceed() error {
	next := this.CurrentFilter.next
	if next != nil {
		this.CurrentFilter = next
		// is this filter applicable?
		if next.IsValid(this.overrider) {
			// TODO replace this with a logger
			//fmt.Printf("===> applying filter with rule '%s'\n", next.rule)
			return next.Apply(this.overrider)
		} else {
			// proceed to next filter
			return this.overrider.Proceed()
		}
		//} else {
		//	http.Error(this.Response, "Page Not Found", http.StatusNotFound)
	}
	// TODO replace this with a logger
	//fmt.Printf("===> unable to proceed from filter with rule '%s'\n", this.CurrentFilter.rule)
	return nil
}

func (this *Context) GetResponse() http.ResponseWriter {
	return this.Response
}

func (this *Context) GetRequest() *http.Request {
	return this.Request
}

func (this *Context) GetSession() ISession {
	if this.Session == nil {
		return nil
	} else {
		return this.Session
	}
}

func (this *Context) SetSession(session ISession) {
	this.Session = session
}

func (this *Context) GetPrincipal() interface{} {
	return this.Principal
}

func (this *Context) SetPrincipal(principal interface{}) {
	this.Principal = principal
}

func (this *Context) GetAttribute(key string) interface{} {
	return this.Attributes[key]
}

func (this *Context) SetAttribute(key string, value interface{}) {
	this.Attributes[key] = value
}

func (this *Context) GetCurrentFilter() *Filter {
	return this.CurrentFilter
}

func (this *Context) SetCurrentFilter(current *Filter) {
	this.CurrentFilter = current
}

type Filterer interface {
	Handle(ctx IContext) error
}

type Filter struct {
	rule        string
	next        *Filter
	handlerFunc func(ctx IContext) error
	handler     Filterer
}

func (this *Filter) IsValid(ctx IContext) bool {
	path := ctx.GetRequest().URL.Path

	if this.rule == "" {
		return true
	} else if strings.HasPrefix(this.rule, "*") {
		return strings.HasSuffix(path, this.rule[1:])
	} else if strings.HasSuffix(this.rule, "*") {
		return strings.HasPrefix(path, this.rule[:len(this.rule)-1])
	} else {
		return path == this.rule
	}
}

func (this *Filter) Apply(ctx IContext) error {
	if this.handlerFunc != nil {
		return this.handlerFunc(ctx)
	} else {
		return this.handler.Handle(ctx)
	}

}

// DO NOT FORGET: filters are applied in reverse order (LIFO)
func NewFilterHandler(contextFactory func(w http.ResponseWriter, r *http.Request) IContext) *FilterHandler {
	this := new(FilterHandler)
	this.contextFactory = contextFactory
	return this
}

type FilterHandler struct {
	first          *Filter
	contextFactory func(w http.ResponseWriter, r *http.Request) IContext
}

func (this *FilterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if this.first != nil {
		var ctx IContext
		if this.contextFactory == nil {
			// default
			ctx = NewContext(w, r)
		} else {
			ctx = this.contextFactory(w, r)
		}
		ctx.SetCurrentFilter(&Filter{next: this.first})
		err := ctx.Proceed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// the last added filter will be the first to be called
func (this *FilterHandler) PushF(rule string, filters ...func(ctx IContext) error) {
	for _, filter := range filters {
		current := &Filter{
			rule:        rule,
			next:        this.first,
			handlerFunc: filter,
		}
		this.first = current
	}
}

// the last added filter will be the first to be called
func (this *FilterHandler) Push(rule string, filters ...Filterer) {
	for _, filter := range filters {
		current := &Filter{
			rule:    rule,
			next:    this.first,
			handler: filter,
		}
		this.first = current
	}
}
