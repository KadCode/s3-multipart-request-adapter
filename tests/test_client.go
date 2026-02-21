package tests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

var client *http.Client

func init() {
	cert, err := tls.LoadX509KeyPair("../certs/cert.pem", "../certs/key.pem")
	if err != nil {
		panic(fmt.Sprintf("failed to load cert/key: %v", err))
	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
		},
		Timeout: 10 * time.Second,
	}
}
