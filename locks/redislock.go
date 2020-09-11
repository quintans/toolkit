package locks

import (
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

type RedisLock struct {
	mu *redsync.Mutex
}

func NewRedisLock(redisAddresses []string, lockName string, expiry time.Duration) (RedisLock, error) {
	pool, err := redisPool(redisAddresses)
	if err != nil {
		return RedisLock{}, err
	}
	lock := redsync.New(pool)
	mu := lock.NewMutex(lockName, redsync.SetExpiry(expiry), redsync.SetTries(2))
	return RedisLock{
		mu: mu,
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
