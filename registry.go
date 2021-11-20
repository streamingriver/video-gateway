package main

import (
	"fmt"
	"sync"
	"time"
)

func NewRegistry() *Registry {
	return &Registry{
		make(map[string]*Item),
		&sync.RWMutex{},
	}
}

type Registry struct {
	registry map[string]*Item
	mu       *sync.RWMutex
}

type Item struct {
	Port string
	Host string
	Seen int64
}

func (r Registry) getURL(ch, file string) *string {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.registry[ch]

	if !ok {
		return nil
	}

	if item.Seen < time.Now().Unix() {
		return nil
	}

	url := fmt.Sprintf("http://%s:%s/%s/%s", item.Host, item.Port, ch, file)
	return &url
}

func (r *Registry) ping(ch, host, port string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.registry[ch]
	if !ok {
		r.registry[ch] = &Item{
			Port: port,
			Host: host,
			Seen: time.Now().Unix() + 3,
		}
	}
	r.registry[ch].Port = port
	r.registry[ch].Host = host
	r.registry[ch].Seen = time.Now().Unix() + 3

}
