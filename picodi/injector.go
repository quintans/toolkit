package picodi

import (
	"errors"
	"reflect"
	"strings"
)

//Provider defines the inerface that a struct must implement
// if it wants to provide for an instance to be wired
type Provider interface {
	Provide() interface{}
}

// Injector is a small framework for Dependency Injection  
type Injector struct {
	providers map[string]interface{}
	instances map[string]interface{}
}

// New creates a new Injector instance
func New() *Injector {
	var inj = new(Injector)
	inj.providers = make(map[string]interface{})
	inj.instances = make(map[string]interface{})

	return inj
}

// Set register a provider.
// This is used like:
// type Foo struct { Bar string }
// inject.Set("foo", Foo{})
// or
// inject.Set("foo", func () Foo {
//   return Foo{}
// })
func (inj *Injector) Set(name string, value interface{}) error {
	if inj.providers[name] != nil || inj.instances[name] != nil {
		return errors.New("There is already a provider or value defined for " + name)
	}

	var t = reflect.TypeOf(value)
	if t.Kind() == reflect.Func {
		inj.providers[name] = value
	} else if _, ok := value.(Provider); ok {
		inj.providers[name] = value
	} else {
		inj.instances[name] = value
	}

	return nil

}

// Get returns the instance by name
func (inj *Injector) Get(name string) (interface{}, error) {
	return inj.get(make([]string, 0), name)
}

func (inj *Injector) get(fetching []string, name string) (interface{}, error) {
	var value = inj.instances[name]
	if value == nil {
		// look for a provider
		var provider = inj.providers[name]
		if provider != nil {
			var value interface{}
			if p, ok := value.(Provider); ok {
				value = p.Provide()
			} else {
				value = callProvider(provider)
			}

			if err := inj.wire(fetching, value); err != nil {
				return nil, err
			}

			inj.instances[name] = value

			return value, nil
		} 

		return nil, errors.New("No provider was found for " + name)
	}

	return value, nil
}

func callProvider(provider interface{}) interface{} {
	var result = reflect.ValueOf(provider).Call([]reflect.Value{})
	return result[0].Interface()
}

// Wire injects dependencies into the instance
func (inj *Injector) Wire(value interface{}) error {
	return inj.wire(make([]string, 0), value)
}

func (inj *Injector) wire(fetching []string, value interface{}) error {
	var val = reflect.ValueOf(value)
	k := val.Kind()
	if k != reflect.Interface && k != reflect.Ptr {
		//return errors.New("Only interfaces or pointers can be wired")
		// Only interfaces or pointers can be wired
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
				if v == name {
					return errors.New("Cyclic wiring: " + strings.Join(fetching, "->"))
				}
			}

			var names = append(fetching, name)

			var fieldValue = s.Field(i)
			if fieldValue.CanSet() {

				var v, err = inj.get(names, name)
				if err != nil {
					return err
				}

				fieldValue.Set(reflect.ValueOf(v))
			} else if method := val.MethodByName("Set" + strings.Title(f.Name)); method.IsValid() {
				// getter defined for the pointer 

				var v, err = inj.get(names, name)
				if err != nil {
					return err
				}

				method.Call([]reflect.Value{reflect.ValueOf(v)})
			}
		}
	}

	return nil
}
