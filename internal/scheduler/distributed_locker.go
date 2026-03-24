package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/DaiYuANg/arcgo/kvx"
	kvlock "github.com/DaiYuANg/arcgo/kvx/module/lock"
	"github.com/go-co-op/gocron/v2"
)

type distributedLocker struct {
	client    kvx.Lock
	keyPrefix string
	ttl       time.Duration
}

func newDistributedLocker(client kvx.Lock, keyPrefix string, ttl time.Duration) gocron.Locker {
	return &distributedLocker{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (l *distributedLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	lockKey := fmt.Sprintf("%s:%s", l.keyPrefix, key)
	kl := kvlock.New(l.client, lockKey, &kvlock.Options{
		TTL:        l.ttl,
		AutoExtend: false,
	})
	if err := kl.Acquire(ctx); err != nil {
		return nil, err
	}
	return &distributedLock{l: kl}, nil
}

type distributedLock struct {
	l *kvlock.Lock
}

func (d *distributedLock) Unlock(ctx context.Context) error {
	return d.l.Release(ctx)
}
