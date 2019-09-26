package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	ServiceURLFlag       = "service-url"
	ListenAddressFlag    = "listen-address"
	ServiceURLDefault    = "https://httpbin.org/headers"
	ListenAddressDefault = ":9000"
)

var (
	serviceURL    string
	listenAddress string
)

func sendError(w http.ResponseWriter, err error) {
	data := make(map[string]interface{})
	data["time"] = time.Now().Format(time.RFC3339)
	data["app"] = "simple-http"
	data["error"] = err.Error()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(data)
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(serviceURL)
	if err != nil {
		sendError(w, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			sendError(w, err)
			return
		}

		data := make(map[string]interface{})
		data["time"] = time.Now().Format(time.RFC3339)
		data["app"] = "simple-http"
		data["response"] = string(body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		json.NewEncoder(w).Encode(data)
		return
	}

	// Add data
	data := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&data)
	data["time"] = time.Now().Format(time.RFC3339)
	data["app"] = "simple-http"

	w.Header().Set("X-App", "simple-http")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	flag.StringVar(&serviceURL, ServiceURLFlag, ServiceURLDefault, "service url to make GET request")
	flag.StringVar(&listenAddress, ListenAddressFlag, ListenAddressDefault, "this server listen address")

	flag.Parse()

	log.Printf("server started with listen address %s with service url %s", listenAddress, serviceURL)

	// start server
	http.HandleFunc("/", handleRequest)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		panic(err)
	}
}
