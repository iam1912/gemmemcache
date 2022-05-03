package gemmemcache

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

type Item struct {
	Val        interface{}
	Expiration time.Time
}

func (i Item) Expire() bool {
	return time.Now().Before(i.Expiration)
}

type MemoryCache struct {
	cache             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stop              chan struct{}
}

func NewMemory(defaultExpiration, cleanupInterval time.Duration) Cache {
	m := &MemoryCache{
		cache:             make(map[string]Item),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}
	go m.StartGC()
	runtime.SetFinalizer(m, stopGC)
	return m
}

func (mc *MemoryCache) Get(key string) interface{} {
	mc.mu.RLock()
	item, ok := mc.cache[key]
	if !ok {
		mc.mu.RUnlock()
		return nil
	}
	mc.mu.RUnlock()
	return item.Val
}

func (mc *MemoryCache) Add(key string, value interface{}, timeout time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if timeout == 0 {
		timeout = mc.defaultExpiration
	}
	mc.cache[key] = Item{
		Val:        value,
		Expiration: time.Now().Add(timeout),
	}
	return nil
}

func (mc *MemoryCache) Delete(key string) error {
	mc.mu.Lock()
	if _, ok := mc.cache[key]; !ok {
		mc.mu.Unlock()
		return errors.New("key not exist")
	}
	delete(mc.cache, key)
	if _, ok := mc.cache[key]; ok {
		mc.mu.Unlock()
		return errors.New("delete key failed")
	}
	mc.mu.Unlock()
	return nil
}

func (mc *MemoryCache) Flush() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cache = make(map[string]Item)
	return nil
}

func (mc *MemoryCache) IsExist(key string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	item, ok := mc.cache[key]
	if ok {
		return item.Expire()
	}
	return false
}

func (mc *MemoryCache) StartGC() {
	ticker := time.NewTicker(mc.cleanupInterval)
	for {
		select {
		case <-ticker.C:
			if keys := mc.expireKeys(); len(keys) != 0 {
				mc.deleteExpiredItem(keys)
			}
		case <-mc.stop:
			ticker.Stop()
			return
		}
	}
}

func (mc *MemoryCache) expireKeys() []string {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	keys := make([]string, 0)
	for key, item := range mc.cache {
		if !item.Expire() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (mc *MemoryCache) deleteExpiredItem(keys []string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, key := range keys {
		delete(mc.cache, key)
	}
}

func stopGC(mc *MemoryCache) {
	mc.stop <- struct{}{}
}

func init() {

}
