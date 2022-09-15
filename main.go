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
	EndpointsMap  map[string]Endpoint `json:"endpointsMap"`
	PollingPeriod time.Duration       `json:"pollingPeriod"`
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
	addrSet map[string]bool
}

func NewPingClient(addrSet map[string]bool) *PingClient {
	return &PingClient{
		addrSet: addrSet,
	}
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

		for epName, endpoint := range config.EndpointsMap {
			if p.addrSet[endpoint.IP] {
				// Skip my endpoint
				continue
			}
			now := time.Now()
			resp, err := http.Get(fmt.Sprintf("http://%s:%s/ping", endpoint.IP, endpoint.Port))
			if err != nil {
				log.Println("error in get request", err)
				continue
			}
			rtt := time.Since(now)
			if resp.StatusCode != http.StatusOK {
				log.Println("error in get request", err)
			}
			log.Printf("rtt to pod %s is %v", epName, rtt)
		}
	}
}

func getLocalIPv4Addresses() (map[string]bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	addrSet := map[string]bool{}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				addrSet[ipnet.IP.String()] = true
			}
		}
	}
	return addrSet, nil
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

	addrSet, err := getLocalIPv4Addresses()
	if err != nil {
		log.Fatalf("failed to fetch interfaces addresses %v", err)
	}

	pongServer := NewPongServer()
	prometheusServer := NewPrometheusServer()
	pingClient := NewPingClient(addrSet)

	go pongServer.ListenAndServe()
	go prometheusServer.ListenAndServe()
	go pingClient.ProbingLoop()

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	pongServer.Shutdown(ctx)
	prometheusServer.Shutdown(ctx)
	cancel()
}
