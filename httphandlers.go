package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Response struct {
	body    []byte
	headers http.Header
	code    int
	err     error
}

type IFetcher interface {
	Fetch(string) *Response
}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

type Fetcher struct {
}

func (f Fetcher) Fetch(url string) *Response {

	hc := http.Client{Timeout: 10 * time.Second}

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", "streaminriveriptv/1.0")

	response, err := hc.Do(request)
	if err != nil {
		return &Response{
			err: err,
		}
	}

	if response.StatusCode/100 != 2 {
		return &Response{
			err:  fmt.Errorf("Invalid status code: %v", response.StatusCode),
			code: response.StatusCode,
		}
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &Response{
			err:  err,
			code: response.StatusCode,
		}
	}
	r := &Response{
		body:    b,
		headers: response.Header.Clone(),
		code:    response.StatusCode,
	}
	return r

}

type HTTPHandler struct {
	registry   *Registry
	Fetcher    IFetcher
	tokens     *Tokens
	transcoder string
}

func (hh *HTTPHandler) ping(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hh.registry.ping(vars["name"], vars["host"], vars["port"])
}

func (hh HTTPHandler) main(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if !hh.tokens.Check(vars["token"], vars["name"], r) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	url := hh.registry.getURL(vars["name"], vars["file"])

	if url == nil {
		response := hh.Fetcher.Fetch(fmt.Sprintf("http://%s/%s/%s", hh.transcoder, vars["name"], vars["file"]))

		for k, v := range response.headers {
			for _, vv := range v {
				w.Header().Set(k, vv)
			}
		}

		w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(response.body)))
		w.Write(response.body)
		return
	}

	response := hh.Fetcher.Fetch(*url)
	if response.err != nil {
		log.Printf("fetch url error: %v", response.err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for k, v := range response.headers {
		for _, vv := range v {
			w.Header().Set(k, vv)
		}
	}
	w.Header().Set("Content-Lenght", fmt.Sprintf("%d", len(response.body)))
	w.Write(response.body)
}
