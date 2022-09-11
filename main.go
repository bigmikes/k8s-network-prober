package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type Config struct {
	EndpointsList []Endpoint    `json:"endpointsList"`
	PollingPeriod time.Duration `json:"pollingPeriod"`
}

type Endpoint struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

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

type PingClient struct {
}

func NewPingClient() *PingClient {
	return &PingClient{}
}

func (p *PingClient) ProbingLoop() {
	pollingPeriod := time.Second * 10
	for {
		time.Sleep(pollingPeriod)
		pathConfigFile := getPathToConfigFile()
		configJSON, err := os.ReadFile(pathConfigFile)
		if err != nil {
			log.Println("error reading file", err)
			continue
		}
		config := Config{}
		err = json.Unmarshal(configJSON, &config)
		if err != nil {
			log.Println("error unmarshalling file", err)
			continue
		}
		pollingPeriod = config.PollingPeriod

		for _, endpoint := range config.EndpointsList {
			now := time.Now()
			resp, err := http.Get(fmt.Sprintf("http://%s:%s/ping", endpoint.IP, endpoint.Port))
			rtt := time.Since(now)
			if resp.StatusCode != http.StatusOK {
				log.Println("error in get request", err)
			}
			log.Println("rtt", rtt)
		}
	}
}

func getPathToConfigFile() string {
	configPath := os.Getenv("NET_PROBER_CONFIG_FILE")
	if configPath == "" {
		configPath = "/etc/netprober/config.json"
	}
	return configPath
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

	pongServer := NewPongServer()
	prometheusServer := NewPrometheusServer()
	pingClient := NewPingClient()

	go pongServer.ListenAndServe()
	go prometheusServer.ListenAndServe()
	go pingClient.ProbingLoop()

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	pongServer.Shutdown(ctx)
	prometheusServer.Shutdown(ctx)
	cancel()
}
