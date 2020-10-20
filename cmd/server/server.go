package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	listenAddress = ":8080"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("host: %s, address: %s, method: %s, requestURI: %s, proto: %s, useragent: %s", r.Host, r.RemoteAddr, r.Method, r.RequestURI, r.Proto, r.UserAgent())
		fmt.Fprintf(w, "Hello World")
	})

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
