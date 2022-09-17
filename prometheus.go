package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func NewPrometheusServer() *PrometheusServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	port := getPrometheusServerPort()
	listenAddr := net.JoinHostPort("", port)

	server := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	return &PrometheusServer{
		mux:    mux,
		server: server,
	}
}

func (p *PrometheusServer) ListenAndServe() {
	err := p.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")

	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

func (p *PrometheusServer) Shutdown(ctx context.Context) {
	log.Println(p.server.Shutdown(ctx))
}

func getPrometheusServerPort() string {
	httpPort := os.Getenv("HTTP_PROMETHEUS_PORT")
	if httpPort == "" {
		httpPort = "2112"
	}
	return httpPort
}
