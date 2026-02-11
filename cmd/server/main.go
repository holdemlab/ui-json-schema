// Package main is the entry point for the ui-json-schema server.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	handler "github.com/holdemlab/ui-json-schema/api"
)

const defaultAddr = ":8080"

func main() {
	addr := defaultAddr
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	registry := handler.NewRegistry()

	h := handler.NewHandler(registry)

	mux := http.NewServeMux()
	mux.HandleFunc("/schema/generate", h.GenerateHandler)

	fmt.Printf("ui-json-schema server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux)) //nolint:gosec // demo server, no TLS needed
}
