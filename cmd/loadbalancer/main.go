package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	mutex    sync.Mutex
	backends []string
)

var currentBackend = 0

func checkHealth(backend string, healthCh chan<- string) {
	healthCh <- backend
}

func isHealthy(server string) bool {
	resp, err := http.Get(server + "/health")
	if err != nil {
		log.Printf("Health check failed for server %s: %v", server, err)
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

var healthyBackends []string

func updateHealthyBackends() {
	healthyCh := make(chan string)

	for {
		for _, backend := range backends {
			go checkHealth(backend, healthyCh)
		}

		select {
		case backend := <-healthyCh:
			// Update the list of healthy backends
			mutex.Lock()
			healthyBackends = append(healthyBackends, backend)
			mutex.Unlock()
		case <-time.After(10 * time.Second):
			// Timeout: clear the list and start over
			mutex.Lock()
			healthyBackends = []string{}
			mutex.Unlock()
		}
	}
}

func selectHealthyBackend() string {
	mutex.Lock()
	defer mutex.Unlock()

	if len(healthyBackends) == 0 {
		return ""
	}
	backend := healthyBackends[currentBackend]
	currentBackend = (currentBackend + 1) % len(healthyBackends)
	return backend
}

func handler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	userAgent := r.UserAgent()
	method := r.Method
	protocol := r.Proto
	accept := r.Header.Get("Accept")

	if accept == "" {
		accept = "* / *"
	}
	forwardReq(w, r)

	fmt.Fprintf(w, "Received request from %s\n%s / %s\nHost: %s\nUser-Agent: %s\nAccept: %s", host, method, protocol, host, userAgent, accept)

}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	switch r.Method {
	case http.MethodPost: // Register a new backend
		if server == "" {
			http.Error(w, "Server not specified", http.StatusBadRequest)
		}

		if !isHealthy(server) {
			http.Error(w, "Backend server is not healthy", http.StatusServiceUnavailable)
			return
		}

		mutex.Lock()
		backends = append(backends, server)
		mutex.Unlock()
		fmt.Fprintf(w, "Registered backend: %s\n", server)

	case http.MethodDelete:
		if server == "" {
			http.Error(w, "Server not specified", http.StatusBadRequest)
		}

		mutex.Lock()
		for i, backend := range backends {
			if backend == server {
				backends = append(backends[:i], backends[i+1:]...)
				break
			}
		}

		mutex.Unlock()
		fmt.Fprintf(w, "Deregistered backend: %s\n", server)

	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func main() {
	go updateHealthyBackends()

	http.HandleFunc("/", handler)
	http.HandleFunc("/register", registerHandler)

	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func forwardReq(w http.ResponseWriter, r *http.Request) {
	backend := selectHealthyBackend()

	// Create a new request to forward
	req, err := http.NewRequest(r.Method, backend, r.Body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}

	// Copy the headers from the original request
	for key, value := range r.Header {
		req.Header.Set(key, value[0])
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request: %v", err)
		http.Error(w, "Error forwarding request:", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Copy the response headers
	for key, value := range resp.Header {
		w.Header().Set(key, value[0])
	}

	// Copy the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body to the client
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}

}
