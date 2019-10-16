package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	serviceURLFlag       = "service-url"
	listenAddressFlag    = "listen-address"
	serviceURLDefault    = "https://httpbin.org/headers"
	listenAddressDefault = ":9000"
)

var (
	serviceURL    string
	listenAddress string
	traceHeaders  = []string{
		"X-Ot-Span-Context",
		"X-Request-Id",
		"X-B3-TraceId",
		"X-B3-SpanId",
		"X-B3-ParentSpanId",
		"X-B3-Sampled",
		"X-B3-Flags",
		"B3",
	}
	respWith503Counter int
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
	code := req.URL.Query().Get("code")
	sleep := req.URL.Query().Get("sleep")
	respWith503 := req.URL.Query().Get("respWith503")

	// if code is not empty, return directly with this code. not call upstream serviceURL
	if code != "" {
		data := make(map[string]interface{})
		data["time"] = time.Now().Format(time.RFC3339)
		data["app"] = "simple-http"

		codeInt, err := strconv.Atoi(code)
		if err != nil {
			codeInt = 200
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(codeInt)
		json.NewEncoder(w).Encode(data)
		return
	}

	// sleep for amount of sleep set
	if sleep != "" {
		sleepDuration, err := time.ParseDuration(sleep)
		if err != nil {
			sleepDuration = time.Second
		}

		time.Sleep(sleepDuration)
	}

	if respWith503 != "" {
		respWith503Count, err := strconv.Atoi(respWith503)
		if err != nil {
			respWith503Count = 5
		}

		// if counter is less then count, response with 503
		if respWith503Counter != respWith503Count {
			respWith503Counter++
			data := make(map[string]interface{})
			data["time"] = time.Now().Format(time.RFC3339)
			data["app"] = "simple-http"
			data["count-503"] = respWith503Counter

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(data)
			return
		}

		// reset to 0
		respWith503Counter = 0
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: transport}

	serviceReq, err := http.NewRequest("GET", serviceURL, nil)
	if err != nil {
		sendError(w, err)
		return
	}

	// propagate trace headers
	for _, h := range traceHeaders {
		hv := req.Header.Get(h)
		if hv == "" {
			continue
		}
		serviceReq.Header.Set(h, hv)
	}

	resp, err := client.Do(serviceReq)
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
		data["count-503"] = respWith503Counter
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
	flag.StringVar(&serviceURL, serviceURLFlag, serviceURLDefault, "service url to make GET request")
	flag.StringVar(&listenAddress, listenAddressFlag, listenAddressDefault, "this server listen address")

	flag.Parse()

	log.Printf("server started with listen address %s with service url %s", listenAddress, serviceURL)

	// start server
	http.HandleFunc("/", handleRequest)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		panic(err)
	}
}
