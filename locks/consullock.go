package locks

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulLockPool struct {
	client *api.Client
}

func NewConsulLockPool(consulAddress string, sessionName string) (ConsulLockPool, error) {
	client, err := api.NewClient(&api.Config{Address: consulAddress})
	if err != nil {
		return ConsulLockPool{}, err
	}

	if err != nil {
		return ConsulLockPool{}, fmt.Errorf("session create err: %v", err)
	}

	return ConsulLockPool{
		client: client,
	}, nil
}

func (p ConsulLockPool) NewLock(lockName string, expiry time.Duration) (ConsulLock, error) {
	return ConsulLock{
		client:   p.client,
		lockName: lockName,
		expiry:   expiry,
	}, nil
}

type ConsulLock struct {
	client   *api.Client
	lockName string
	expiry   time.Duration
	done     chan struct{}
}

func (l ConsulLock) Lock() (chan struct{}, error) {
	sEntry := &api.SessionEntry{
		TTL:       l.expiry.String(),
		LockDelay: 1 * time.Millisecond,
		Behavior:  "delete",
	}
	sID, _, err := l.client.Session().Create(sEntry, nil)
	if err != nil {
		return nil, err
	}

	// auto renew session
	l.done = make(chan struct{})
	go func() {
		err = l.client.Session().RenewPeriodic(sEntry.TTL, sID, nil, l.done)
		if err != nil {
			close(l.done)
		}
	}()

	acquireKv := &api.KVPair{
		Session: sID,
		Key:     l.lockName,
		Value:   []byte(sID),
	}
	acquired, _, err := l.client.KV().Acquire(acquireKv, nil)
	if err != nil {
		return nil, err
	}

	if !acquired {
		return nil, fmt.Errorf("Unable to acquire lock for key %s", l.lockName)
	}

	return l.done, nil
}

func (l ConsulLock) Unlock() error {
	close(l.done)
	_, err := l.client.KV().Delete(l.lockName, nil)
	return err
}
