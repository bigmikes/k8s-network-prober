package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Define Prometheus metrics.
var (
	labels          = []string{"dest_pod_name", "dest_ip"}
	interPodLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "netprober_interpod_latency_s",
		Help:    "Histogram of inter-Pod HTTP probe latencies",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 25),
	}, labels)
)

// Config is the network-prober configuration.
type Config struct {
	EndpointsMap  map[string]Endpoint `json:"endpointsMap"`
	PollingPeriod time.Duration       `json:"pollingPeriod"`
}

// Endpoint is an IPv4 address and TCP port.
type Endpoint struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
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
