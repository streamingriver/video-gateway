package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/avarf/getenvs"
)

func main() {

	tokens := NewTokens(getenvs.GetEnvString("TOKENS_URL", "http://sr-admin-gui/api/tokens"))
	tokens.Worker()

	router := mux.NewRouter()

	handlers := new(HTTPHandler)
	handlers.registry = NewRegistry()
	handlers.tokens = tokens

	router.HandleFunc("/register/backend/{name}/{host}/{port}", handlers.ping)

	router.HandleFunc("/{token}/{name}/{file:.*}", handlers.main)

	log.Printf("Starting service on %s", getenvs.GetEnvString("SUPER_CONFIG_PORT", ":80"))
	log.Fatal(http.ListenAndServe(getenvs.GetEnvString("SUPER_CONFIG_PORT", ":80"), router))

}
