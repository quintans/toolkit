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
		mu: mu,
	}
}

type RedisLock struct {
	mu *redsync.Mutex
}

func (l RedisLock) Lock() (bool, error) {
	err := l.mu.Lock()
	if err == redsync.ErrFailed {
		return false, nil
	}
	return err == nil, err
}

func (l RedisLock) Extend() (bool, error) {
	return l.mu.Extend()
}

func (l RedisLock) Unlock() (bool, error) {
	return l.mu.Unlock()
}
