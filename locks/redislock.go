package locks

import (
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

type RedisLockPool struct {
	lock *redsync.Redsync
}

func NewRedisLockPool(redisAddresses []string) (RedisLockPool, error) {
	pool, err := redisPool(redisAddresses)
	if err != nil {
		return RedisLockPool{}, err
	}
	return RedisLockPool{
		lock: redsync.New(pool),
	}, nil
}

func redisPool(addrs []string) ([]redsync.Pool, error) {
	pool := make([]redsync.Pool, len(addrs))
	for k, v := range addrs {
		p := &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", v)
			},
		}
		pool[k] = p
	}
	return pool, nil
}

func (p RedisLockPool) NewLock(lockName string, expiry time.Duration) RedisLock {
	mu := p.lock.NewMutex(lockName, redsync.SetExpiry(expiry), redsync.SetTries(2))
	return RedisLock{
		mu:        mu,
		heartbeat: expiry / 2,
	}
}

type RedisLock struct {
	mu        *redsync.Mutex
	heartbeat time.Duration
	done      chan struct{}
}

func (l RedisLock) Lock() (chan struct{}, error) {
	err := l.mu.Lock()
	if err == redsync.ErrFailed {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.done = make(chan struct{})
	go func() {
		ticker := time.NewTicker(l.heartbeat)
		defer ticker.Stop()
		for {
			select {
			case <-l.done:
				l.Unlock()
				return
			case <-ticker.C:
				ok, _ := l.mu.Extend()
				if !ok {
					l.Unlock()
					return
				}
			}
		}

	}()

	return l.done, nil
}

func (l RedisLock) Unlock() (bool, error) {
	close(l.done)
	return l.mu.Unlock()
}
