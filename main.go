package main

import (
	"bytes"
	"crypto/tls"
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

var client *http.Client

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

func DoRequest(endpoint DeepLXEndpoint, body []byte) (*http.Response, error) {
	// copy request
	req, err := http.NewRequest("POST", endpoint.Url+"/translate", bytes.NewBuffer(body))
	if err != nil {
		log.Println("Failed to create request:", err)
		return nil, err
	}
	// add auth header
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Ch-Ua", "\"Chromium\";v=\"94\", \"Microsoft Edge\";v=\"94\", \";Not A Brand\";v=\"99\"")
	req.Header.Add("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Add("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Authority", "www.deepl.com")
	req.Header.Add("Pragma", "no-cache")
	if endpoint.Token != "" {
		req.Header.Add("Authorization", "Bearer "+endpoint.Token)
	}
	// send request
	return client.Do(req)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")                                              // 允许任何源
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")               // 允许的方法
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With") // 允许的标头
}

func HelloworldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "DeepLx Load Balancer v0.2, https://github.com/nerdneilsfield/deeplx-load-balancer")
}

func LoadBalancerHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if config.Token != "" {
		// check auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}
		if authHeader != "Bearer "+config.Token {
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
	}

	endpoints := make([]DeepLXEndpoint, len(config.Endpoints))
	copy(endpoints, config.Endpoints)

	var resp *http.Response
	var err error

	// 读取请求
	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request", http.StatusInternalServerError)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	for len(endpoints) > 0 {
		endpoint := RandomEndpoint(endpoints)
		resp, err = DoRequest(endpoint, bodyBytes)

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
	w.WriteHeader(http.StatusOK)
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

	defaultCipherSuites := []uint16{0xc02f, 0xc030, 0xc02b, 0xc02c, 0xcca8, 0xcca9, 0xc013, 0xc009,
		0xc014, 0xc00a, 0x009c, 0x009d, 0x002f, 0x0035, 0xc012, 0x000a}
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				CipherSuites: append(defaultCipherSuites[8:], defaultCipherSuites[:8]...),
				MaxVersion:   tls.VersionTLS12,
			},
			ForceAttemptHTTP2: false,
		},
	}

	// Start balancer
	// StartBalancer(config)
	http.HandleFunc("/", HelloworldHandler)
	http.HandleFunc("/translate", LoadBalancerHandler)

	log.Fatal(http.ListenAndServe(":1188", nil))
}
