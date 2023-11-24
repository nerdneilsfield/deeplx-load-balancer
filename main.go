package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type DeepLXEndpoint struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type DeepLXBalancerConfig struct {
	Token     string           `json:"token"`
	Endpoints []DeepLXEndpoint `json:"endpoints"`
}

var configFile = flag.String("config", "config.json", "Path to config file")

var config = DeepLXBalancerConfig{}

func LoadConfig(path string) {
	// open config file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// read config file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded config with %d endpoints", len(config.Endpoints))
	for _, endpoint := range config.Endpoints {
		log.Printf("Endpoint: %s", endpoint.Url)
	}
}

func main() {
	flag.Parse()

	// Load config
	LoadConfig(*configFile)

	// Start balancer
	// StartBalancer(config)
}
