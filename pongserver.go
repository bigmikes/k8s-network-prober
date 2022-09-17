package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// PongServer is a simple HTTP server that replies
// on the /ping endpoint with a "pong" response.
type PongServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func NewPongServer() *PongServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handleGetPing)

	port := getPongServerPort()
	listenAddr := net.JoinHostPort("", port)

	server := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	return &PongServer{
		mux:    mux,
		server: server,
	}
}

func (p *PongServer) ListenAndServe() {
	err := p.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")

	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

func (p *PongServer) Shutdown(ctx context.Context) {
	log.Println(p.server.Shutdown(ctx))
}

func handleGetPing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping request")
	io.WriteString(w, "pong\n")
}

func getPongServerPort() string {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	return httpPort
}
