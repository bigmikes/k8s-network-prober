package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// PingClient loops periodically over the list
// of endpoints and measure the HTTP RTT.
type PingClient struct {
	// addrSet is the set of local endpoints
	// to skip during the probing loop
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
			// Send a request and measure the RTT
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
			// Save the RTT in Prometheus metrics
			interPodLatency.WithLabelValues(epName, endpoint.IP).Observe(rtt.Seconds())
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
