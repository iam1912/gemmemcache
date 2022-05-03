package gemmemcache

import (
	"errors"
	"time"
)

type Cache interface {
	Get(key string) interface{}
	Add(key string, val interface{}, timeout time.Duration) error
	Delete(key string) error
	Flush() error
	IsExist(key string) bool
	StartGC()
}

type Instance func() Cache

var adapters = make(map[string]Instance)

func Register(name string, instance Instance) error {
	if instance == nil {
		return errors.New("instance is empty")
	}
	if _, ok := adapters[name]; ok {
		return errors.New("instance is exist")
	}
	adapters[name] = instance
	return nil
}

func New(name string) (Cache, error) {
	c, ok := adapters[name]
	if !ok {
		return nil, errors.New("instance is not exist of name")
	}
	cache := c()
	return cache, nil
}
