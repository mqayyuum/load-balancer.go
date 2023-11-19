package main

import (
	"flag"
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	userAgent := r.UserAgent()
	method := r.Method
	protocol := r.Proto
	accept := r.Header.Get("Accept")
	if accept == "" {
		accept = "* / *"
	}
	fmt.Println("Incoming request from", string(host))
	fmt.Fprintf(w, "Received request from %s\n%s / %s\nHost: %s\nUser-Agent: %s\nAccept: %s", host, method, protocol, host, userAgent, accept)
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	port := flag.String("port", "8081", "Port to run the server on")
	flag.Parse()

	http.HandleFunc("/health", health)
	http.HandleFunc("/", handler)

	fmt.Printf("Server is running on http://localhost:%s\n", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
