package gemmemcache

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileItem struct {
	Val        interface{}
	Expiration time.Time
}

type FileCache struct {
	cachePath         string
	fileSuffix        string
	diretoryLevel     int
	defaultExpiration time.Duration
	mu                sync.RWMutex
}

func NewFileCache(path, suffix string, level int, defaultExpiration time.Duration) Cache {
	fc := &FileCache{
		cachePath:         path,
		fileSuffix:        suffix,
		diretoryLevel:     level,
		defaultExpiration: defaultExpiration,
	}
	fc.initFile()
	return fc
}

func (fc *FileCache) initFile() {
	if ok, _ := fc.exist(fc.cachePath); !ok {
		os.MkdirAll(fc.cachePath, os.ModePerm)
	}
}

func (fc *FileCache) Get(key string) interface{} {
	name := fc.splicingFileName(key)
	fc.mu.RLock()
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fc.mu.RUnlock()
		return nil
	}
	fc.mu.RUnlock()
	var item FileItem
	err = GobDecode(data, &item)
	if err != nil {
		return nil
	}
	if !time.Now().Before(item.Expiration) {
		return nil
	}
	return item.Val
}

func (fc *FileCache) Add(key string, value interface{}, timeout time.Duration) error {
	gob.Register(value)

	if timeout == 0 {
		timeout = fc.defaultExpiration
	}
	item := &FileItem{Val: value, Expiration: time.Now().Add(timeout)}
	data, err := GobEncode(item)
	if err != nil {
		return err
	}
	name := fc.splicingFileName(key)
	fc.mu.Lock()
	defer fc.mu.Unlock()
	err = ioutil.WriteFile(name, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (fc *FileCache) Delete(key string) error {
	name := fc.splicingFileName(key)
	if ok, _ := fc.exist(name); !ok {
		return errors.New("key is not exist")
	}
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if err := os.Remove(name); err != nil {
		return err
	}
	return nil
}

func (fc *FileCache) Flush() error {
	return nil
}

func (fc *FileCache) IsExist(key string) bool {
	ok, _ := fc.exist(fc.splicingFileName(key))
	return ok
}

func (fc *FileCache) StartGC() {}

func (fc *FileCache) splicingFileName(key string) string {
	hash := md5.New()
	hash.Write([]byte(key))
	md5Key := hex.EncodeToString(hash.Sum(nil))
	path := fc.cachePath
	switch fc.diretoryLevel {
	case 1:
		path = filepath.Join(path, md5Key[0:3])
	case 2:
		path = filepath.Join(path, md5Key[0:3], md5Key[4:7])
	}
	if ok, _ := fc.exist(path); !ok {
		os.MkdirAll(path, os.ModePerm)
	}
	return filepath.Join(path, fmt.Sprintf("%s%s", md5Key, fc.fileSuffix))
}

func (fc *FileCache) exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GobEncode(val interface{}) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(val)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func GobDecode(data []byte, item *FileItem) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(&item)
}

func init() {

}
