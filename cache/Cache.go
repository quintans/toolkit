package cache

type Cache interface {
	GetIfPresent(key string) interface{}
	Delete(key string)
	Get(key string, callback func() interface{}) interface{}
	Put(key string, value interface{})
}
