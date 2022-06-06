package redislock

import (
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

type Pool struct {
	lock *redsync.Redsync
}

func NewPool(redisAddresses []string) (Pool, error) {
	pool, err := redisPool(redisAddresses)
	if err != nil {
		return Pool{}, err
	}
	return Pool{
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

func (p Pool) NewLock(lockName string, expiry time.Duration) Lock {
	mu := p.lock.NewMutex(lockName, redsync.SetExpiry(expiry), redsync.SetTries(2))
	return Lock{
		mu:        mu,
		heartbeat: expiry / 2,
	}
}

type Lock struct {
	mu        *redsync.Mutex
	heartbeat time.Duration
	done      chan struct{}
}

func (l Lock) Lock() (chan struct{}, error) {
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
				l.mu.Unlock()
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

func (l Lock) Unlock() error {
	close(l.done)
	_, err := l.mu.Unlock()
	return err
}
