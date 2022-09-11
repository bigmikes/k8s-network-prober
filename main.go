package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PingPongServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func NewPingPongServer() *PingPongServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handleGetPing)

	port := getPingPongServerPort()
	listenAddr := net.JoinHostPort("", port)

	server := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	return &PingPongServer{
		mux:    mux,
		server: server,
	}
}

func (p *PingPongServer) ListenAndServe() {
	err := p.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")

	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

func (p *PingPongServer) Shutdown(ctx context.Context) {
	log.Println(p.server.Shutdown(ctx))
}

func handleGetPing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping request")
	io.WriteString(w, "pong\n")
}

func getPingPongServerPort() string {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	return httpPort
}

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

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			dat, err := os.ReadFile("/etc/netprober/endpointsJSON")
			if err != nil {
				log.Println("error reading file", err)
			} else {
				log.Println(string(dat))
			}
			time.Sleep(time.Second * 10)
		}
	}()

	pingPongServer := NewPingPongServer()
	prometheusServer := NewPrometheusServer()

	go pingPongServer.ListenAndServe()
	go prometheusServer.ListenAndServe()

	<-c

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	pingPongServer.Shutdown(ctx)
	prometheusServer.Shutdown(ctx)
}
