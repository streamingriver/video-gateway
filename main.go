package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gitlab.com/avarf/getenvs"
)

var (
	flagHelp = flag.Bool("h", false, "")
	exit     = make(chan struct{})
)

func main() {
	flag.Parse()
	if *flagHelp {
		fmt.Println(`FFMPEG_SERVER_HOST=video-transcoding
TOKENS_URL=http://sr-admin-gui/api/tokens
SUPER_CONFIG_PORT=:80`)
		return
	}

	tokens := NewTokens(getenvs.GetEnvString("TOKENS_URL", "http://sr-admin-gui/api/tokens"))
	tokens.Worker()

	router := mux.NewRouter()

	handlers := new(HTTPHandler)
	handlers.registry = NewRegistry()
	handlers.tokens = tokens
	handlers.Fetcher = NewFetcher()
	handlers.transcoder = getenvs.GetEnvString("FFMPEG_SERVER_HOST", "video-transcoding")

	router.HandleFunc("/reload", handlers.reload)
	router.HandleFunc("/register/backend/{name}/{host}/{port}", handlers.ping)

	router.HandleFunc("/{token}/{name}/{file:.*}", handlers.main)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	srv := &http.Server{
		Addr:    getenvs.GetEnvString("SUPER_CONFIG_PORT", ":80"),
		Handler: router,
	}
	defer func() {
		// extra handling here
		cancel()
	}()

	go func() {
		if getenvs.GetEnvString("NO_OUTPUT", "") == "" {
			log.Printf("Starting service on %s", getenvs.GetEnvString("SUPER_CONFIG_PORT", ":80"))
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-exit
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
