package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	CachePath       = "cache"
	FileCacheSuffix = ".cache"
)

type Cache struct {
	sync.RWMutex
	items      map[string]Item
	maxEntries int
	usingDisk  bool
}

type Item struct {
	Value      interface{}
	Expiration int64
	Create     time.Time
}

func New(maxEntries int) *Cache {
	items := make(map[string]Item)
	cache := Cache{
		items:      items,
		maxEntries: maxEntries,
		usingDisk:  false,
		numEntries: 0,
	}
	return &cache
}

func (c *Cache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if !c.usingDisk {
		c.items[key] = Item{
			Value: value,
		}
		if len(c.items) > c.maxEntries {
			c.usingDisk = true
		}
	} else {
		c.SaveToFile(key, value)
	}
}

func (c *Cache) Get(key string, value interface{}) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	var item Item
	if !c.usingDisk {
		item, ok := c.items[key]
		if !ok {
			return nil, false
		}
	} else {
		item, err := c.LoadFromFile(key)
		if err != nil {
			return nil, false
		}
	}

	return item.Value, true
}

func (c *Cache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	if !c.usingDisk {
		if _, ok := c.items[key]; !ok {
			return errors.New("No such key")
		}

		delete(c.items, key)
	} else {
		return c.DeleteFile(key)
	}
	return nil
}

func (c *Cache) SaveFromMemory() error {

}

func (c *Cache) LoadFromDisk() error {

}

func (c *Cache) SaveToFile(key string, value interface{}) error {
	fp := filepath.Join(CachePath, fmt.Sprintf("%s%s", key, FileCacheSuffix))
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(value)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fp, buf.Bytes(), 0644)
	// if _, err := os.Stat(fp); err == nil {
	// 	// exists
	// } else if os.IsNotExist(err) {
	// 	// doesnot exist
	// }
	return err
}

func (c *Cache) LoadFromFile(key string) (Item, error) {
	fp := filepath.Join(CachePath, fmt.Sprintf("%s%s", key, FileCacheSuffix))
	data, err := ioutil.ReadFile(fp)
	var item Item
	if err != nil {
		return item, err
	}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&item)
	if err != nil {
		return item, err
	}
	return item, nil
}

func (c *Cache) DeleteFile(key string) error {
	err := os.Remove(filepath.Join(CachePath, fmt.Sprintf("%s%s", key, FileCacheSuffix)))
	if err != nil {
		return err
	}
	files, _ := ioutil.ReadDir(CachePath)

	if len(files) <= c.maxEntries {
		c.usingDisk = false
		c.LoadFromDisk()
	}
	return nil
}
