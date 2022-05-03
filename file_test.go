package gemmemcache

import (
	"os"
	"testing"
	"time"
)

func TestFile(t *testing.T) {
	fc := NewFileCache("./cache", ".bin", 2, time.Minute*3)
	if err := fc.Add("key1", "value1", time.Hour*3); err != nil {
		t.Error("add error")
	}
	if val := fc.Get("key1"); val == nil || val.(string) != "value1" {
		t.Error("get error")
	}
	if err := fc.Add("key1", "value2", time.Hour*3); err != nil {
		t.Error("update error")
	}
	if !fc.IsExist("key1") {
		t.Error("check error")
	}
	if val := fc.Get("key1"); val == nil || val.(string) != "value2" {
		t.Error("get error")
	}
	if err := fc.Delete("key1"); err != nil {
		t.Error("delete error")
	}
	if fc.IsExist("key1") {
		t.Error("check error")
	}
	os.RemoveAll("./cache")
}
