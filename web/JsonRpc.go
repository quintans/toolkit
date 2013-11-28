package web

import (
	"encoding/json"
	"fmt"
	tk "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/log"
	"io/ioutil"
	"net/http"
	"reflect"
	"unicode"
)

var logger = log.LoggerFor("github.com/quintans/toolkit/web")

// Makes a struct responsible for handling json-rpc calls.
// The endpoint will be composet by the service name and action name. ex order/item
// The rules for the action parameters are:
// * can have at most two parameters
// * if it has two parameters, the first must be of the type web.IContext
//
// The rules for the action return values are:
// * can have at most two return values
// * if it has two parameters, the last must be of the type error
//
// valid signature:  MyStruct.MyAction([web.IContext][any]) [any][error]

const (
	CALL        = "_CALL_"
	UNKNOWN_SRV = "JSONRPC01"
	UNKNOWN_ACT = "JSONRPC02"
)

type Action struct {
	hasContext  bool
	payloadType reflect.Type
	filter      *Filter
}

func (this *Action) PushFilterFunc(filters ...func(ctx IContext) error) {
	for _, filter := range filters {
		current := &Filter{
			next:        this.filter,
			handlerFunc: filter,
		}
		this.filter = current
	}
}

func (this *Action) PushFilter(filters ...Filterer) {
	for _, filter := range filters {
		current := &Filter{
			next:    this.filter,
			handler: filter,
		}
		this.filter = current
	}
}

type Service struct {
	instance reflect.Value
	actions  map[string]*Action
}

func (this *Service) GetAction(actionName string) *Action {
	action, ok := this.actions[actionName]
	if !ok {
		panic("The action " + actionName + " was not found in service")
	}
	return action
}

// the last added filter will be the first to be called
func (this *Service) PushFilterFunc(actionName string, filters ...func(ctx IContext) error) {
	action := this.GetAction(actionName)
	action.PushFilterFunc(filters...)
}

// the last added filter will be the first to be called
func (this *Service) PushFilter(actionName string, filters ...Filterer) {
	action := this.GetAction(actionName)
	action.PushFilter(filters...)
}

type JsonRpc struct {
	services       map[string]*Service
	contextFactory func(w http.ResponseWriter, r *http.Request) IContext
}

func NewJsonRpc(contextFactory func(w http.ResponseWriter, r *http.Request) IContext) *JsonRpc {
	this := new(JsonRpc)
	this.services = make(map[string]*Service)
	this.contextFactory = contextFactory
	return this
}

func (this *JsonRpc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var ctx IContext
	if this.contextFactory == nil {
		// default
		ctx = NewContext(w, r)
	} else {
		ctx = this.contextFactory(w, r)
	}
	err := this.Handle(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// implementation of web.Filterer interface
func (this *JsonRpc) Handle(ctx IContext) error {
	w := ctx.GetResponse()
	r := ctx.GetRequest()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Expires", "-1")

	uri := ctx.GetRequest().RequestURI
	var service string
	var action string
	last := len(uri)
	for i := last - 1; i > 0; i-- {
		if uri[i] == '/' {
			if action == "" {
				action = uri[i+1 : last]
			} else if service == "" {
				service = uri[i+1 : last]
				break
			}
			last = i
		}
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	svc, ok := this.services[service]
	if !ok {
		return &tk.Fail{UNKNOWN_SRV, service}
	}

	act, ok := svc.actions[action]
	if !ok {
		return &tk.Fail{UNKNOWN_ACT, fmt.Sprintf("%s.%s", service, action)}
	}

	mthd := svc.instance.MethodByName(action)
	call := &invokation{
		action:  act,
		method:  mthd,
		payload: payload,
	}
	ctx.SetAttribute(CALL, call)
	ctx.SetCurrentFilter(act.filter)
	// call filter
	err = act.filter.Apply(ctx)

	if err != nil {
		return err
	}

	if call.response != nil {
		w.Write(call.response)
	}

	return nil
}

type invokation struct {
	action   *Action
	method   reflect.Value
	payload  []byte
	response []byte
}

func invokeFilter(ctx IContext) error {
	call := ctx.GetAttribute(CALL).(*invokation)

	var err error
	call.response, err = invoke(ctx, call.action, call.method, call.payload)
	return err
}

func (this *JsonRpc) Register(svc interface{}) *Service {
	return this.RegisterAs("", svc)
}

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()    // interface type
	contextType = reflect.TypeOf((*IContext)(nil)).Elem() // interface type
)

func (this *JsonRpc) RegisterAs(name string, svc interface{}) *Service {
	typ := reflect.TypeOf(svc)
	if typ.Kind() != reflect.Ptr {
		panic("Supplied instance must be a pointer.")
	}

	// Only structs are supported
	if typ.Elem().Kind() != reflect.Struct {
		panic("Supplied instance is not a struct.")
	}

	actions := make(map[string]*Action)

	var serviceName string
	if name == "" {
		serviceName = typ.Name()
	} else {
		serviceName = name
	}

	// loop through the struct's fields and set the map
	for i := 0; i < typ.NumMethod(); i++ {
		p := typ.Method(i)
		if isExported(p.Name) {
			action := new(Action)

			logger.Debugf("Registering JSON-RPC %s/%s", serviceName, p.Name)

			// validate argument types
			size := p.Type.NumIn()
			if size > 3 {
				panic(fmt.Sprintf("Invalid service %s.%s. Service actions can only have at the most two  parameters.",
					typ.Elem().Name(), p.Name))
			} else if size > 2 {
				t := p.Type.In(1)
				if t != contextType {
					panic(fmt.Sprintf("Invalid service %s.%s. In a two paramater action the first must be the interface web.IContext.",
						typ.Elem().Name(), p.Name))
				}
			}

			if size == 3 {
				action.payloadType = p.Type.In(2)
				action.hasContext = true
			} else if size == 2 {
				t := p.Type.In(1)
				if t != contextType {
					action.payloadType = t
				} else {
					action.hasContext = true
				}
			}

			//logger.Debugf("Has Contex: %t; Payload Type: %s", action.hasContext, action.payloadType)

			// validate return types
			size = p.Type.NumOut()
			if size > 2 {
				panic(fmt.Sprintf("Invalid service %s.%s. Service actions can only have at the most two return values.",
					typ.Elem().Name(), p.Name))
			} else if size > 1 && errorType != p.Type.Out(1) {
				panic(fmt.Sprintf("Invalid service %s.%s. In a two return values actions the second can only be an error. Found %s.",
					typ.Elem().Name(), p.Name))
			}

			action.filter = &Filter{handlerFunc: invokeFilter}
			actions[p.Name] = action
		}
	}

	val := reflect.ValueOf(svc)
	s := &Service{val, actions}
	this.services[serviceName] = s
	return s
}

func invoke(ctx IContext, act *Action, m reflect.Value, args []byte) ([]byte, error) {
	var param reflect.Value
	var err error
	if act.payloadType != nil {
		// get pointer
		param = reflect.New(act.payloadType)
		// TODO: what happens if args is "null" ???
		err = json.Unmarshal(args, param.Interface())
		if err != nil {
			logger.Errorf("An error ocurred when unmarshalling the call for %s\n\tinput: %s\n\terror: %s", ctx.GetRequest().URL.Path, args, err)
			return nil, err
		}
	}
	params := make([]reflect.Value, 0)
	if act.hasContext {
		params = append(params, reflect.ValueOf(ctx))
	}
	if act.payloadType != nil {
		params = append(params, param.Elem())
	}

	results := m.Call(params)

	// check for error
	var result []byte
	for k, v := range results {
		if v.Type() == errorType {
			if !v.IsNil() {
				return nil, v.Interface().(error)
			}
			break
		} else {
			// stores the result to return at the end of the check
			data := results[k].Interface()
			result, err = json.Marshal(data)
			if err != nil {
				logger.Errorf("An error ocurred when marshalling the response from %s\n\tresponse: %v\n\terror: %s", ctx.GetRequest().URL.Path, data, err)
				return nil, err
			}
		}
	}

	return result, nil
}

func isExported(name string) bool {
	return unicode.IsUpper(rune(name[0]))
}

func capitalize(str string) string {
	var s string
	if len(str) > 0 {
		s = string(unicode.ToUpper(rune(str[0])))
	}
	if len(str) > 1 {
		s += str[1:]
	}
	return s
}
