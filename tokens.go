package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func NewTokens(url string, other ...interface{}) *Tokens {
	t := &Tokens{
		fetchFrom: url,
		mu:        &sync.RWMutex{},
		workerMu:  &sync.RWMutex{},
		maps:      &T{make(map[string]bool), make(map[string]string), make(map[string]string)},
		work:      false,
	}
	return t
}

type Tokens struct {
	fetchFrom string
	mu        *sync.RWMutex
	workerMu  *sync.RWMutex
	maps      *T
	work      bool
}

type T struct {
	ADDR map[string]bool   `json:"addr"`
	IP   map[string]string `json:"ip"`
	CH   map[string]string `json:"ch"`
}

// Check if user match conditions and can watch
func (t Tokens) Check(token, channel string, req *http.Request) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	host := req.Header.Get("X-Real-IP")
	if strings.Contains(req.Header.Get("X-Real-IP"), ":") {
		var err error
		host, _, err = net.SplitHostPort(req.Header.Get("X-Real-IP"))
		if err != nil {
			log.Printf("%v", err)
			return false
		}
	}

	ip, ok := t.maps.IP[token]
	if !ok {
		return false
	}
	if ip != host {
		return false
	}
	ch, ok := t.maps.CH[token]
	if !ok {
		return false
	}
	if ch != channel {
		return false
	}

	return true
}

// Worker periodicaly fetches tokens from api
func (t *Tokens) Worker() {
	t.workerMu.Lock()
	defer t.workerMu.Unlock()
	t.work = true
	go t.updateTokens()
}

func (t *Tokens) StopWorker() {
	t.workerMu.Lock()
	defer t.workerMu.Unlock()
	t.work = false
}

func (t *Tokens) updateTokens() {
	for {
		t.workerMu.RLock()
		if !t.work {
			t.workerMu.RUnlock()
			break
		}
		t.workerMu.RUnlock()
		t.tokens()
		time.Sleep(2 * time.Second)

	}
}

func (t *Tokens) tokens() {
	body := t.remoteGet(t.fetchFrom)
	if body == nil {
		return
	}
	t.mu.Lock()
	err := json.Unmarshal(body, t.maps)
	t.mu.Unlock()
	if err != nil {
		log.Printf("updateTokens error: %v", err)
		return
	}
}

func (t *Tokens) remoteGet(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		log.Printf("remote_get error: %v", err)
		return nil
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("remote_get error: %v", err)
		return nil
	}

	return b
}
