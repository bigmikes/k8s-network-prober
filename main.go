package main

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping request")
	io.WriteString(w, "pong\n")
}

func main() {
	http.HandleFunc("/ping", getPing)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	listenAddr := net.JoinHostPort("", httpPort)

	log.Println("Starting HTTP server on", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")

	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}
