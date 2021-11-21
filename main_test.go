package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

type TestFetcher struct {
	t *testing.T
}

func (tf *TestFetcher) Fetch(url string) *Response {
	// tf.t.Logf("%v", url)
	r := &Response{body: []byte("response success")}
	return r
}

func requestPing(registry *Registry, url string) {
	req := httptest.NewRequest(http.MethodGet, "/register/backend/kanalas/hostas/portas", nil)
	w := httptest.NewRecorder()
	handlers := new(HTTPHandler)
	handlers.registry = registry
	handlers.Fetcher = &TestFetcher{}

	host, port, err := net.SplitHostPort(url)
	if err != nil {
		panic(err)
	}

	vars := map[string]string{
		"name": "kanalas",
		"host": host,
		"port": port,
	}
	req = mux.SetURLVars(req, vars)

	handlers.ping(w, req)

	res := w.Result()
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
}

func TestFetch(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "labas")
	}))
	defer svr.Close()

	// t.Logf("%s", NewFetcher().Fetch(svr.URL).body)
}

func TestHTTPPing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/register/backend/kanalas/hostas/portas", nil)
	w := httptest.NewRecorder()
	handlers := new(HTTPHandler)
	handlers.registry = NewRegistry()
	handlers.Fetcher = &TestFetcher{t}

	vars := map[string]string{
		"name": "kanalas",
		"host": "hostas",
		"port": "80",
	}

	req = mux.SetURLVars(req, vars)

	handlers.ping(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "" {
		t.Errorf("expected empty response got %v", string(data))
	}

	item, ok := handlers.registry.registry["kanalas"]
	if !ok {
		t.Errorf("expected %s channel to be set", "kanalas")
	}
	if item.Host != "hostas" {
		t.Errorf("expected 'hostas' got %v", item.Host)
	}
	if item.Port != "80" {
		t.Errorf("expected 'portas' got %v", item.Port)
	}
}

func TestHTTPMain(t *testing.T) {
	svr := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "labas")
	}))
	svr.Config.Addr = "127.0.0.1:65000"
	svr.Start()
	defer svr.Close()

	tokensServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		{"addr":{"192.168.1.1":true},"ip":{"tokenas":"192.168.1.1"},"ch":{"tokenas":"kanalas"}}
		`)
	}))
	tokensServer.Start()
	defer tokensServer.Close()

	registry := NewRegistry()
	handlers := new(HTTPHandler)
	handlers.registry = registry
	handlers.Fetcher = &TestFetcher{t}
	handlers.tokens = NewTokens(tokensServer.URL)
	handlers.tokens.tokens()
	requestPing(registry, strings.ReplaceAll(svr.URL, "http://", ""))

	req := httptest.NewRequest(http.MethodGet, "/tokenas/kanalas/failas.ts", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()

	///{token}/{name}/{file:.*}
	vars := map[string]string{
		"token": "tokenas",
		"name":  "kanalas",
		"file":  "failas.ts",
	}

	req = mux.SetURLVars(req, vars)

	req.Host = svr.URL

	handlers.main(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "response success" {
		t.Errorf("expected 'response success' response got %v", string(data))
	}

}

func TestHTTPMainRealFetcher(t *testing.T) {
	svr := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "labas")
	}))
	svr.Config.Addr = "127.0.0.1:65000"
	svr.Start()
	defer svr.Close()

	tokensServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		{"addr":{"192.168.1.1":true},"ip":{"tokenas":"192.168.1.1"},"ch":{"tokenas":"kanalas"}}
		`)
	}))
	tokensServer.Start()
	defer tokensServer.Close()

	registry := NewRegistry()
	handlers := new(HTTPHandler)
	handlers.registry = registry
	handlers.Fetcher = NewFetcher()
	handlers.tokens = NewTokens(tokensServer.URL)
	handlers.tokens.tokens()
	requestPing(registry, strings.ReplaceAll(svr.URL, "http://", ""))

	req := httptest.NewRequest(http.MethodGet, "/tokenas/kanalas/failas.ts", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()

	///{token}/{name}/{file:.*}
	vars := map[string]string{
		"token": "tokenas",
		"name":  "kanalas",
		"file":  "failas.ts",
	}

	req = mux.SetURLVars(req, vars)

	req.Host = svr.URL

	handlers.main(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "labas" {
		t.Errorf("expected 'labas' response got %v", string(data))
	}

}

func TestTokenUnauthorized(t *testing.T) {
	svr := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "labas")
	}))
	svr.Config.Addr = "127.0.0.1:65000"
	svr.Start()
	defer svr.Close()

	tokensServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		{"addr":{"192.168.1.1":true},"ip":{"tokenas":"192.168.1.1"},"ch":{"tokenas":"kanalas"}}
		`)
	}))
	tokensServer.Start()
	defer tokensServer.Close()

	registry := NewRegistry()
	handlers := new(HTTPHandler)
	handlers.registry = registry
	handlers.Fetcher = NewFetcher()
	handlers.tokens = NewTokens(tokensServer.URL)
	handlers.tokens.tokens()
	requestPing(registry, strings.ReplaceAll(svr.URL, "http://", ""))

	req := httptest.NewRequest(http.MethodGet, "/tokenas/kanalas/failas.ts", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")
	w := httptest.NewRecorder()

	///{token}/{name}/{file:.*}
	vars := map[string]string{
		"token": "kazkoks",
		"name":  "kanalas",
		"file":  "failas.ts",
	}

	req = mux.SetURLVars(req, vars)

	req.Host = svr.URL

	handlers.main(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if strings.Trim(string(data), "\n") != "Unauthorized" {
		t.Errorf("expected 'Unauthorized' response got %v", string(data))
	}

}

func TestTokens(t *testing.T) {
	svr := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		{"addr":{"192.168.1.1":true},"ip":{"token":"192.168.1.1"},"ch":{"token":"7f587efec2912ea2"}}
		`)
	}))
	svr.Start()
	defer svr.Close()

	tokens := NewTokens(svr.URL + "/api/tokens")

	tokens.tokens()

	if tokens.maps.IP["token"] != "192.168.1.1" {
		t.Errorf("IP for 'token' should be '192.168.1' got %v", tokens.maps.IP["token"])
	}

}

func TestTokensWorker(t *testing.T) {
	tokensServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		{"addr":{"192.168.1.1":true},"ip":{"tokenas":"192.168.1.1"},"ch":{"tokenas":"kanalas"}}
		`)
	}))
	tokensServer.Start()
	defer tokensServer.Close()

	tokens := NewTokens(tokensServer.URL)
	tokens.Worker()

	time.Sleep(time.Second * 1)

	tokens.StopWorker()

	time.Sleep(time.Second * 3)

	if tokens.work != false {
		t.Errorf("worker should stopper already")
	}
}
