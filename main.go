package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
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

func RandomEndpoint(endpoints []DeepLXEndpoint) DeepLXEndpoint {
	// pick a random endpoint
	seed := time.Now().UnixNano()
	rander := rand.New(rand.NewSource(seed))
	index := rander.Intn(len(endpoints))
	return endpoints[index]
}

func removeEndpoint(endpoints []DeepLXEndpoint, endpoint DeepLXEndpoint) []DeepLXEndpoint {
	for i, e := range endpoints {
		if e == endpoint {
			return append(endpoints[:i], endpoints[i+1:]...)
		}
	}
	return endpoints
}

func DoRequest(endpoint DeepLXEndpoint, token string, r *http.Request) (*http.Response, error) {
	// copy request
	req, err := http.NewRequest("POST", endpoint.Url+"/translate", r.Body)
	if err != nil {
		log.Println("Failed to create request:", err)
		return nil, err
	}
	// add auth header
	for k, v := range r.Header {
		req.Header[k] = v
	}
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	// send request
	return http.DefaultClient.Do(req)
}

func HelloworldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "DeepLx Load Balancer v0.1, https://github.com/nerdneilsfield/deeplx-load-balancer")
}

func LoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	endpoints := make([]DeepLXEndpoint, len(config.Endpoints))
	copy(endpoints, config.Endpoints)

	var resp *http.Response
	var err error

	for len(endpoints) > 0 {
		endpoint := RandomEndpoint(endpoints)
		resp, err = DoRequest(endpoint, config.Token, r)

		if err != nil || resp.StatusCode != 200 {
			// 从 endpoints 列表中移除失败的 endpoint
			log.Println("Failed to request endpoint:", endpoint.Url)
			log.Println("Error:", err)
			if resp != nil {
				log.Println("Status:", resp.StatusCode)
			}
			endpoints = removeEndpoint(endpoints, endpoint)
			continue
		}

		// 成功响应
		break
	}

	if resp == nil || resp.StatusCode != 200 {
		// 所有 endpoints 都失败了
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"data": "Service unavailable"}`)
		// http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// 成功响应
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		// 处理写入错误
		log.Println("Failed to write response:", err)
	}
}

func main() {
	flag.Parse()

	// Load config
	LoadConfig(*configFile)

	// Start balancer
	// StartBalancer(config)
	http.HandleFunc("/", HelloworldHandler)
	http.HandleFunc("/translate", LoadBalancerHandler)

	log.Fatal(http.ListenAndServe(":1188", nil))
}
