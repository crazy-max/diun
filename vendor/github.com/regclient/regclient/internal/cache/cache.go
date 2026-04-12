//go:build go1.18

// Package cache is used to store values with limits.
// Items are automatically pruned when too many entries are stored, or values become stale.
package cache

import (
	"sort"
	"sync"
	"time"

	"github.com/regclient/regclient/types/errs"
)

type Cache[k comparable, v any] struct {
	mu       sync.Mutex
	minAge   time.Duration
	maxAge   time.Duration
	minCount int
	maxCount int
	timer    *time.Timer
	entries  map[k]*Entry[v]
}

type Entry[v any] struct {
	used  time.Time
	value v
}

type sortKeys[k comparable] struct {
	keys   []k
	lessFn func(a, b k) bool
}

type conf struct {
	minAge   time.Duration
	maxCount int
}

type cacheOpts func(*conf)

func WithAge(age time.Duration) cacheOpts {
	return func(c *conf) {
		c.minAge = age
	}
}

func WithCount(count int) cacheOpts {
	return func(c *conf) {
		c.maxCount = count
	}
}

func New[k comparable, v any](opts ...cacheOpts) Cache[k, v] {
	c := conf{}
	for _, opt := range opts {
		opt(&c)
	}
	maxAge := c.minAge + (c.minAge / 10)
	minCount := 0
	if c.maxCount > 0 {
		minCount = int(float64(c.maxCount) * 0.9)
	}
	return Cache[k, v]{
		minAge:   c.minAge,
		maxAge:   maxAge,
		minCount: minCount,
		maxCount: c.maxCount,
		entries:  map[k]*Entry[v]{},
	}
}

func (c *Cache[k, v]) Delete(key k) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	if len(c.entries) == 0 && c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

func (c *Cache[k, v]) Set(key k, val v) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = &Entry[v]{
		used:  time.Now(),
		value: val,
	}
	if len(c.entries) > c.maxCount {
		c.pruneLocked()
	} else if c.timer == nil {
		// prune resets the timer, so this is only needed if the prune wasn't triggered
		c.timer = time.AfterFunc(c.maxAge, c.prune)
	}
}

func (c *Cache[k, v]) Get(key k) (v, error) {
	if c == nil {
		var val v
		return val, errs.ErrNotFound
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok {
		if e.used.Add(c.minAge).Before(time.Now()) {
			// entry expired
			go c.prune()
		} else {
			c.entries[key].used = time.Now()
			return e.value, nil
		}
	}
	var val v
	return val, errs.ErrNotFound
}

func (c *Cache[k, v]) prune() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pruneLocked()
}

func (c *Cache[k, v]) pruneLocked() {
	// sort key list by last used date
	keyList := make([]k, 0, len(c.entries))
	for key := range c.entries {
		keyList = append(keyList, key)
	}
	sk := sortKeys[k]{
		keys: keyList,
		lessFn: func(a, b k) bool {
			return c.entries[a].used.Before(c.entries[b].used)
		},
	}
	sort.Sort(&sk)
	// prune entries
	now := time.Now()
	cutoff := now.Add(c.minAge * -1)
	nextTime := now
	delCount := len(keyList) - c.minCount
	for i, key := range keyList {
		if i < delCount || c.entries[key].used.Before(cutoff) {
			delete(c.entries, key)
		} else {
			nextTime = c.entries[key].used
			break
		}
	}
	// set next timer
	if len(c.entries) > 0 {
		dur := nextTime.Sub(now) + c.maxAge
		if c.timer == nil {
			// this shouldn't be possible
			c.timer = time.AfterFunc(dur, c.prune)
		} else {
			c.timer.Reset(dur)
		}
	} else if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}

func (sk *sortKeys[k]) Len() int {
	return len(sk.keys)
}

func (sk *sortKeys[k]) Less(i, j int) bool {
	return sk.lessFn(sk.keys[i], sk.keys[j])
}

func (sk *sortKeys[k]) Swap(i, j int) {
	sk.keys[i], sk.keys[j] = sk.keys[j], sk.keys[i]
}
