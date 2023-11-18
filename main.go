package main

import (
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
