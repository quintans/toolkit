package cache

import (
	"sync"
	"time"
)

type ExpirationCache struct {
	items    map[string]*item
	timeout  time.Duration
	interval time.Duration
	sync.Mutex
}

type item struct {
	value      interface{}
	expiration time.Time
}

// Returns true if the item has expired.
func (i *item) expired() bool {
	return i.expiration.Before(time.Now())
}

func NewExpirationCache(timeout time.Duration, interval time.Duration) *ExpirationCache {
	cache := new(ExpirationCache)
	cache.items = make(map[string]*item)
	cache.timeout = timeout
	cache.interval = interval
	go cache.cleanup()
	return cache
}

func (this *ExpirationCache) cleanup() {
	tick := time.Tick(this.interval)
	for {
		<-tick
		this.deleteExpired()
	}
}

// Delete all expired items from the cache.
func (this *ExpirationCache) deleteExpired() {
	this.Lock()
	defer this.Unlock()
	for k, v := range this.items {
		if v.expired() {
			delete(this.items, k)
		}
	}
}

func (this *ExpirationCache) GetIfPresentAndTouch(key string) interface{} {
	this.Lock()
	defer this.Unlock()

	v, ok := this.items[key]
	if ok {
		v.expiration = time.Now().Add(this.timeout)
		return v.value
	}
	return nil
}

func (this *ExpirationCache) GetIfPresent(key string) interface{} {
	this.Lock()
	v, ok := this.items[key]
	this.Unlock()
	if ok {
		return v.value
	}
	return nil
}

func (this *ExpirationCache) Delete(key string) {
	this.Lock()
	delete(this.items, key)
	this.Unlock()
}

func (this *ExpirationCache) Get(key string, callback func() interface{}) interface{} {
	return this.GetWithDuration(key, callback, this.timeout)
}

func (this *ExpirationCache) GetWithDuration(key string, callback func() interface{}, duration time.Duration) interface{} {
	this.Lock()
	defer this.Unlock()

	v, ok := this.items[key]
	if !ok {
		v = &item{callback(), time.Now().Add(this.timeout)}
		this.items[key] = v
	}
	return v.value
}

func (this *ExpirationCache) Put(key string, value interface{}) {
	this.PutWithDuration(key, value, this.timeout)
}

// put a value in the cache, overwriting any previous value for that key
func (this *ExpirationCache) PutWithDuration(key string, value interface{}, duration time.Duration) {
	this.Lock()
	// defer now sice I do not know what will happen in a out of memory error
	defer this.Unlock()
	this.items[key] = &item{value, time.Now().Add(duration)}
}

func (this *ExpirationCache) Touch(key string) {
	this.TouchWithDuration(key, this.timeout)
}

func (this *ExpirationCache) TouchWithDuration(key string, duration time.Duration) {
	this.Lock()
	v, ok := this.items[key]
	if ok {
		v.expiration = time.Now().Add(duration)
	}
	this.Unlock()
}
