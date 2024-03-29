package consullock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

type Pool struct {
	client *api.Client
}

func NewPool(consulAddress string) (Pool, error) {
	api.DefaultConfig()
	client, err := api.NewClient(&api.Config{Address: consulAddress})
	if err != nil {
		return Pool{}, err
	}

	if err != nil {
		return Pool{}, fmt.Errorf("session create err: %v", err)
	}

	return Pool{
		client: client,
	}, nil
}

func (p Pool) NewLock(lockName string, expiry time.Duration) *Lock {
	return &Lock{
		client:   p.client,
		lockName: lockName,
		expiry:   expiry,
	}
}

type Lock struct {
	client   *api.Client
	sID      string
	lockName string
	expiry   time.Duration
	done     chan struct{}
	mu       sync.Mutex
}

func (l *Lock) Lock(ctx context.Context) (chan struct{}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.done != nil {
		return nil, fmt.Errorf("this lock '%s' is already acquire. Unlock it first", l.lockName)
	}

	sEntry := &api.SessionEntry{
		TTL:      l.expiry.String(),
		Behavior: "delete",
	}
	options := &api.WriteOptions{}
	options = options.WithContext(ctx)
	sID, _, err := l.client.Session().Create(sEntry, options)
	if err != nil {
		return nil, err
	}

	l.sID = sID
	acquireKv := &api.KVPair{
		Session: l.sID,
		Key:     l.lockName,
		Value:   []byte(sID),
	}
	acquired, _, err := l.client.KV().Acquire(acquireKv, options)
	if err != nil {
		return nil, err
	}

	if !acquired {
		_, _ = l.client.Session().Destroy(sID, options)
		return nil, nil
	}

	// auto renew session
	l.done = make(chan struct{})
	go func() {
		// we use a new options because context may no longer be usable
		err := l.client.Session().RenewPeriodic(sEntry.TTL, sID, &api.WriteOptions{}, l.done)
		if err != nil {
			_ = l.Unlock(options.Context())
		}
	}()

	return l.done, nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.done == nil {
		return nil
	}

	close(l.done)
	l.done = nil

	return nil
}

func (l *Lock) WaitForUnlock(ctx context.Context) error {
	opts := &api.QueryOptions{}
	opts = opts.WithContext(ctx)

	done := make(chan error, 1)
	heartbeat := l.expiry / 2

	go func() {
		ticker := time.NewTicker(heartbeat)
		defer ticker.Stop()
		for {
			kv, _, err := l.client.KV().Get(l.lockName, opts)
			if err != nil {
				done <- err
				return
			}
			if kv == nil {
				done <- nil
				return
			}
			<-ticker.C
		}
	}()
	err := <-done

	return err
}
