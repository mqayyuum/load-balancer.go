package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}
var currentBackend = 0

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

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func forwardReq(w http.ResponseWriter, r *http.Request) {
	backend := backends[currentBackend]

	currentBackend = (currentBackend + 1) % len(backends)

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
