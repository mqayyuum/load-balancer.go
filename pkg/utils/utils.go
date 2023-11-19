package utils

import (
	"log"
	"net"
	"net/http"
)

type Req struct {
	IP        string
	UserAgent string
	Method    string
	Protocol  string
	Accept    string
	Host      string
}

func GetUtils(r *http.Request) (Req, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Fatalf("Error getting IP address: %v", err)
		return Req{}, err
	}

	accept := r.Header.Get("Accept")
	if accept == "" {
		accept = "* / *"
	}

	return Req{
		IP:        ip,
		UserAgent: r.UserAgent(),
		Method:    r.Method,
		Accept:    accept,
		Host:      r.Host,
	}, nil
}
