package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/mqayyuum/load-balancer-go/pkg/utils"
)

func handler(w http.ResponseWriter, r *http.Request) {
	info, err := utils.GetUtils(r)
	if err != nil {
		http.Error(w, "Unable to parse request", http.StatusBadRequest)
	}
	fmt.Println("Incoming request from", string(info.IP))
	fmt.Fprintf(w, "Received request from %s\n%s / %s\nHost: %s\nUser-Agent: %s\nAccept: %s", info.IP, info.Method, info.Protocol, info.Host, info.UserAgent, info.Accept)
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
