package picodi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// IWire is an interface for any implementation that has to implement wiring.
//
// With the advent of plugins (Go 1.8), DI might be usefull.
// We could ask the plugin to wire itself with the supplied Depency PicoDI.
// This way context could evolve independently in the main program and in the plugins.
type IWire interface {
	Wire(interface{}) error
	Get(string) (interface{}, error)
	Set(string, func() interface{})
	SetValue(string, interface{})
}

// PicoDI is a small framework for Dependency Injection.
// With this we can concentrate all the configuration in one place and avoid the import cycles problem.
// I also see it being useful in plugins.
type PicoDI struct {
	providers map[string]func() interface{}
	instances map[string]interface{}
}

// check if interface is fully implemented
var _ IWire = &PicoDI{}

type chain struct {
	field string
	name  string
}

func (c chain) String() string {
	return c.field + " \"" + c.name + "\""
}

// New creates a new PicoDI instance
func New() *PicoDI {
	var pdi = new(PicoDI)
	pdi.providers = make(map[string]func() interface{})
	pdi.instances = make(map[string]interface{})

	return pdi
}

// Set register a provider.
// This is used like:
// PicoDI.Set("foo", func () interface{} {
//   return Foo{}
// })
// If the returned value of the provider is to be wired, it must return a pointer
func (pdi *PicoDI) Set(name string, fn func() interface{}) {
	pdi.providers[name] = fn
}

// SetNamedValue registers a value with a name.
// Internally it will register a provider that returns value.
// This is used like:
// type Foo struct { Bar string }
// PicoDI.SetNamedValue("Foo", Foo{})
// If the returned value of the provider is to be wired, it must return a pointer
func (pdi *PicoDI) SetValue(name string, value interface{}) {
	pdi.Set(name, func() interface{} { return value })
}

// Get returns the instance by name
func (pdi *PicoDI) Get(name string) (interface{}, error) {
	return pdi.get(make([]chain, 0), name)
}

func (pdi *PicoDI) get(fetching []chain, name string) (interface{}, error) {
	var value = pdi.instances[name]
	if value == nil {
		// look for a provider
		var provider = pdi.providers[name]
		if provider != nil {
			var value = provider()

			if err := pdi.wire(fetching, value); err != nil {
				return nil, err
			}

			pdi.instances[name] = value

			return value, nil
		}

		return nil, errors.New("No provider was found for " + joinChain(fetching))
	}

	return value, nil
}

func callProvider(provider interface{}) interface{} {
	var result = reflect.ValueOf(provider).Call([]reflect.Value{})
	return result[0].Interface()
}

// Wire injects dependencies into the instance.
// Dependencies marked for wiring without name will be mapped to their type name.
func (pdi *PicoDI) Wire(value interface{}) error {
	return pdi.wire(make([]chain, 0), value)
}

func (pdi *PicoDI) wire(fetching []chain, value interface{}) error {
	var val = reflect.ValueOf(value)
	k := val.Kind()
	if k != reflect.Interface && k != reflect.Ptr {
		if len(fetching) == 0 {
			// the first wiring must be valid
			var err = fmt.Sprintf("The first wiring must be a interface or a pointer: %s", value)
			return errors.New(err)
		}
		// Other than interfaces and pointers will ne ignored
		// because they cannot can be wired
		return nil
	}

	// gets the inner struct
	var s = val.Elem()

	// struct type
	var t = s.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if name, ok := f.Tag.Lookup("wire"); ok {
			if name == "" {
				//return errors.New("Undeclared injection name for " + f.Name)
				name = f.Type.Name()
			}

			// see if this name is already being fetched
			for _, v := range fetching {
				if v.name == name {
					return errors.New("Cyclic wiring: " + joinChain(fetching))
				}
			}

			var fieldName = t.Name() + "." + f.Name
			var names = append(fetching, chain{fieldName, name})

			var fieldValue = s.Field(i)
			if fieldValue.CanSet() {

				var v, err = pdi.get(names, name)
				if err != nil {
					return err
				}

				fieldValue.Set(reflect.ValueOf(v))
			} else if method := val.MethodByName("Set" + strings.Title(f.Name)); method.IsValid() {
				// getter defined for the pointer

				var v, err = pdi.get(names, name)
				if err != nil {
					return err
				}

				method.Call([]reflect.Value{reflect.ValueOf(v)})
			}
		}
	}

	return nil
}

func joinChain(fetching []chain) string {
	var s = make([]string, len(fetching))
	for k, v := range fetching {
		s[k] = v.String()
	}

	return strings.Join(s, "->")
}
