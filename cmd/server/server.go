package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	listenAddress = ":8080"
)

func checkStatusCode(code int, codes []int) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}

	return false
}

func main() {
	reqs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Name:      "requests_total",
			Help:      "Total requests by HTTP result code and method.",
		},
		[]string{"code", "method"})

	prometheus.MustRegister(reqs)

	var statusCodes = []int{200, 200, 200, 200, 200, 400, 500, 502, 503}

	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("host: %s, address: %s, method: %s, requestURI: %s, proto: %s, useragent: %s", r.Host, r.RemoteAddr, r.Method, r.RequestURI, r.Proto, r.UserAgent())
		fmt.Fprintf(w, "Hello World")
	})
	router.HandleFunc("/status", promhttp.InstrumentHandlerCounter(reqs, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("host: %s, address: %s, method: %s, requestURI: %s, proto: %s, useragent: %s", r.Host, r.RemoteAddr, r.Method, r.RequestURI, r.Proto, r.UserAgent())
		status, ok := r.URL.Query()["status"]
		if !ok || len(status[0]) < 1 || status[0] == "random" {
			index := rand.Intn(len(statusCodes))
			w.WriteHeader(statusCodes[index])
			return
		}

		s, err := strconv.Atoi(status[0])
		if err == nil && checkStatusCode(s, statusCodes) {
			w.WriteHeader(s)
			return
		}

		w.WriteHeader(400)
	})))
	router.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	go func() {
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)

		<-term
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatalf("Failed to shutdown server gracefully: %s", err.Error())
		}

		log.Printf("Shutdown server...")
		os.Exit(0)
	}()

	log.Printf("Server listen on: %s", listenAddress)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server died unexpected: %s", err.Error())
	}
}
