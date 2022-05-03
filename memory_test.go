package gemmemcache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	bc := NewMemory(time.Minute*3, time.Minute*5)
	if err := bc.Add("key1", 1, 0); err != nil {
		t.Error("add error")
	}
	if val := bc.Get("key1"); val == nil || val.(int) != 1 {
		t.Error("get error")
	}
	bc.Add("key1", 2, time.Hour*3)
	if val := bc.Get("key1"); val == nil || val.(int) != 2 {
		t.Error("get error")
	}
	if !bc.IsExist("key1") {
		t.Error("check error")
	}
	if err := bc.Delete("key1"); err != nil {
		t.Error("delete error")
	}
	if bc.IsExist("key1") {
		t.Error("check error")
	}
	if val := bc.Get("key1"); val != nil {
		t.Error("get error")
	}
}
