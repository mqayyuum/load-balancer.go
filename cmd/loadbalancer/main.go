package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mqayyuum/load-balancer-go/pkg/utils"
)

var (
	mutex           sync.Mutex
	backends        []string
	healthyBackends []string
)

var currentBackend = 0

type HealthStatus struct {
	Address string
	Healthy bool
}

func checkHealth(backend string, healthCh chan<- HealthStatus) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Get(backend + "/health")
	if err != nil {
		log.Printf("Health check failed for server %s: %v\n", backend, err)
		healthCh <- HealthStatus{Address: backend, Healthy: false}
		return
	}

	defer resp.Body.Close()

	health := resp.StatusCode == http.StatusOK
	healthCh <- HealthStatus{Address: backend, Healthy: health}
	if !health {
		log.Printf("Server %s is not healthy. Status Code %d\n", backend, resp.StatusCode)
	}
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

func updateHealthyBackends() {
	healthyCh := make(chan HealthStatus)

	for {
		for _, backend := range backends {
			go checkHealth(backend, healthyCh)
		}

		for range backends {
			status := <-healthyCh
			mutex.Lock()
			if status.Healthy {
				found := false
				for _, b := range healthyBackends {
					if b == status.Address {
						found = true
						break
					}
				}
				if !found {
					healthyBackends = append(healthyBackends, status.Address)
				}

			} else {
				for i, backend := range healthyBackends {
					if backend == status.Address {
						healthyBackends = append(healthyBackends[:i], healthyBackends[i+1:]...)
						break
					}
				}
			}
			mutex.Unlock()
		}
		time.Sleep(10 * time.Second)
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
	info, err := utils.GetUtils(r)
	if err != nil {
		http.Error(w, "Unable to process request", http.StatusBadRequest)
	}
	forwardReq(w, r)

	fmt.Fprintf(w, "Received request from %s\n%s / %s\nHost: %s\nUser-Agent: %s\nAccept: %s", info.IP, info.Method, info.Protocol, info.Host, info.UserAgent, info.Accept)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	switch r.Method {
	case http.MethodGet: // Get list of registered backend
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(backends); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

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

	case http.MethodDelete: // Delete a registered backend
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

	log.Printf("Current backend is: %s\n", backend)

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
