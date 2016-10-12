package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	tk "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/log"
)

var filtersLog = log.LoggerFor("github.com/quintans/toolkit/web/filters")

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

	Payload(value interface{}) error
	PathVars(value interface{}) error
	QueryVars(value interface{}) error
	Vars(value interface{}) error
	Reply(value interface{}) error
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	this := new(Context)
	this.Init(w, r)
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
	jsonQuery     string
}

func (this *Context) Init(w http.ResponseWriter, r *http.Request) {
	this.Response = w
	this.Request = r
	this.Attributes = make(map[string]interface{})
}

// Proceed proceeds to the next valid rule
func (this *Context) Proceed() error {
	var next = this.CurrentFilter.next
	if next != nil {
		if next.rule == "" {
			this.CurrentFilter = next
			filtersLog.Debug("executing filter without rule")
			return next.Apply(this)
		} else {
			// go to the next valid filter.
			// I don't use recursivity for this, because it can be very deep
			for n := this.CurrentFilter.next; n != nil; n = n.next {
				if n.rule != "" && n.IsValid(this) {
					this.CurrentFilter = n
					filtersLog.Debugf("executing filter %s", n.rule)
					return n.Apply(this)
				}
			}
		}
	}

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

func (this *Context) Payload(value interface{}) error {
	if this.Request.Body != nil {
		payload, err := ioutil.ReadAll(this.Request.Body)
		if err != nil {
			return err
		}

		return json.Unmarshal(payload, value)
	}

	return nil
}

func (this *Context) PathVars(value interface{}) error {
	var filter = this.GetCurrentFilter()
	if filter.jsonPath != "" {
		return json.Unmarshal([]byte(filter.jsonPath), value)
	}

	return nil
}

func (this *Context) QueryVars(value interface{}) error {
	var t = reflect.TypeOf(value)
	if t.Kind() != reflect.Ptr {
		return errors.New("Query value must be a pointer, not " + t.Kind().String())
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("Query value must be of kind struct, not " + t.Kind().String())
	}

	var values = this.GetRequest().URL.Query()
	if this.jsonQuery == "" && len(values) > 0 {
		var json = ""

		for i := 0; i < t.NumField(); i++ {
			var f = t.Field(i)
			var name = f.Name
			var v = values.Get(name)

			if v == "" {
				name = tk.UncapFirst(f.Name)
				v = values.Get(name)
			}

			if v != "" {
				var k = f.Type.Kind()
				if k == reflect.Bool {
					v = toJsonVal(v, "bool")
				} else if k >= reflect.Int && k <= reflect.Float64 {
					v = toJsonVal(v, "number")
				} else {
					v = toJsonVal(v, "string")
				}

				if len(json) > 0 {
					json += ", "
				}
				json += fmt.Sprintf(`"%s": %s`, name, v)
			}
		}
		if json != "" {
			this.jsonQuery = "{" + json + "}"
		}
	}

	if this.jsonQuery != "" {
		return json.Unmarshal([]byte(this.jsonQuery), value)
	}

	return nil
}

func (this *Context) Vars(value interface{}) error {
	if err := this.QueryVars(value); err != nil {
		return err
	}
	if err := this.PathVars(value); err != nil {
		return err
	}

	return nil
}

func (this *Context) Reply(value interface{}) error {
	result, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = this.Response.Write(result)

	return err
}

type Filterer interface {
	Handle(ctx IContext) error
}

type Filter struct {
	rule           string
	template       []string
	jsonPath       string
	allowedMethods []string

	next    *Filter
	handler func(ctx IContext) error
}

func (this *Filter) IsValid(ctx IContext) bool {
	// verify if method is allowed
	var allowed bool
	if this.allowedMethods == nil {
		allowed = true
	} else {
		var method = ctx.GetRequest().Method
		if method == "" {
			method = "GET"
		}
		for _, v := range this.allowedMethods {
			if method == v {
				allowed = true
				break
			}
		}
	}

	if allowed {
		var path = ctx.GetRequest().URL.Path
		if strings.HasPrefix(this.rule, "*") {
			return strings.HasSuffix(path, this.rule[1:])
		} else if strings.HasSuffix(this.rule, "*") {
			return strings.HasPrefix(path, this.rule[:len(this.rule)-1])
		} else if this.template != nil {
			var ok bool
			this.jsonPath, ok = this.parseToJson(path)
			return ok
		} else {
			return path == this.rule
		}
	}

	return false
}

// parseToJson converts path vars into json string
// and also checks if its a valid match with the url template
func (this *Filter) parseToJson(path string) (string, bool) {
	var json = ""
	var parts = strings.Split(path, "/")

	if len(parts) != len(this.template) {
		return "", false
	}

	for k, v := range this.template {
		if strings.HasPrefix(v, "{") {
			var name = v[1 : len(v)-1]
			var nameType = strings.Split(name, ":")
			name = nameType[0]
			var typ string
			if len(nameType) > 1 {
				typ = nameType[1]
			}

			var val = toJsonVal(parts[k], typ)
			if len(json) > 0 {
				json += ", "
			}
			json += fmt.Sprintf(`"%s": %s`, name, val)

		} else if v != parts[k] {
			return "", false
		}
	}

	return "{" + json + "}", true
}

func toJsonVal(ori string, typ string) string {
	var val = ori

	switch typ {
	case "number":
	case "bool":
		if val == "1" || val == "true" || val == "t" {
			val = "true"
		} else {
			val = "false"
		}
	default:
		val = "\"" + val + "\""
	}

	return val
}

func (this *Filter) Apply(ctx IContext) error {
	return this.handler(ctx)
}

func (this *Filter) Push(filters ...func(ctx IContext) error) *Filter {
	for _, f := range filters {
		last := this.fetchEmpty("")
		last.handler = f
		return last
	}

	return this
}

func (this *Filter) fetchEmpty(rule string) *Filter {
	// if there is no handler returns self
	if this.handler == nil {
		this.rule = rule
		return this
	}

	// goes through the chain stoping at the last link
	var last = this
	for next := last.next; next != nil; {
		last = next
	}

	// apend a new Filter
	last.next = &Filter{
		rule: rule,
	}

	// returns appended filter
	return last.next
}

// Join joins several filters in to one.
// This way we can have sevral filters under one rule.
func Join(filters ...func(ctx IContext) error) func(ctx IContext) error {
	var filter = new(Filter)
	filter.Push(filters...)

	return func(ctx IContext) error {
		var last = ctx.GetCurrentFilter()
		ctx.SetCurrentFilter(&Filter{next: filter})
		var err = ctx.Proceed()
		ctx.SetCurrentFilter(last)
		return err
	}
}

func NewFilterHandler(contextFactory func(w http.ResponseWriter, r *http.Request) IContext) *FilterHandler {
	this := new(FilterHandler)
	this.contextFactory = contextFactory
	return this
}

type FilterHandler struct {
	first          *Filter
	last           *Filter
	contextFactory func(w http.ResponseWriter, r *http.Request) IContext
	lastRule       string
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

func (this *FilterHandler) GET(rule string, filters ...func(ctx IContext) error) {
	this.PushMethod([]string{"GET"}, rule, filters...)
}

func (this *FilterHandler) POST(rule string, filters ...func(ctx IContext) error) {
	this.PushMethod([]string{"POST"}, rule, filters...)
}

func (this *FilterHandler) PUT(rule string, filters ...func(ctx IContext) error) {
	this.PushMethod([]string{"PUT"}, rule, filters...)
}

func (this *FilterHandler) DELETE(rule string, filters ...func(ctx IContext) error) {
	this.PushMethod([]string{"DELETE"}, rule, filters...)
}

func (this *FilterHandler) Push(rule string, filters ...func(ctx IContext) error) {
	this.PushMethod(nil, rule, filters...)
}

// PushMethod adds the filters to the end of the last filters.
// If the rule does NOT start with '/' the applied rule will be
// the concatenation of the last rule that started with '/' and ended with a '*'
// with this one (the '*' is omitted).
// ex: /greet/* + sayHi/{Id} = /greet/sayHi/{Id}
func (this *FilterHandler) PushMethod(methods []string, rule string, filters ...func(ctx IContext) error) {
	if strings.HasPrefix(rule, "/") && strings.HasSuffix(rule, "*") {
		this.lastRule = rule[:len(rule)-1]
	} else if !strings.HasPrefix(rule, "/") {
		if this.lastRule == "" {
			rule = "/" + rule
		} else {
			rule = this.lastRule + rule
		}
	}

	for k, filter := range filters {
		current := &Filter{
			handler:        filter,
			allowedMethods: methods,
		}
		// rule is only set for the first filter
		if k == 0 {
			current.rule = rule
			if i := strings.Index(rule, "{"); i != -1 {
				current.template = strings.Split(rule, "/")
			}
		}

		if this.first == nil {
			this.first = current
		} else {
			this.last.next = current
		}
		this.last = current
	}
}
