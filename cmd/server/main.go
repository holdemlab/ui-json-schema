// Package main is the entry point for the ui-json-schema server.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	handler "github.com/holdemlab/ui-json-schema/api"
	"github.com/holdemlab/ui-json-schema/parser"
	"github.com/holdemlab/ui-json-schema/schema"
)

const defaultAddr = ":8080"

func main() {

	ui, _ := parser.GenerateUISchemaWithOptions(Form{}, schema.DefaultOptions())
	jsonData, _ := json.MarshalIndent(ui, "", "  ")
	fmt.Printf("Generated UI Schema: %s\n", jsonData)
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

type Form struct {
	Name  string `json:"name" form:"category=Personal"`
	Email string `json:"email" form:"category=Personal"`
	Role  string `json:"role" form:"category=Work"`
	Bio   string `json:"bio" form:"category=Work;multiline"`
}
