package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/holdemlab/ui-json-schema/parser"
	"github.com/holdemlab/ui-json-schema/schema"
)

// maxRequestBody limits the request body size (2 MB).
const maxRequestBody = 2 << 20

// generateRequest represents the incoming request for schema generation.
type generateRequest struct {
	// Type is the registered Go type name. If set, JSON payload is ignored.
	Type string `json:"type,omitempty"`
	// Data is a raw JSON object to generate schemas from.
	Data json.RawMessage `json:"data,omitempty"`
}

// generateResponse is the response containing both schemas.
type generateResponse struct {
	Schema   *schema.JSONSchema      `json:"schema"`
	UISchema *schema.UISchemaElement `json:"uischema"`
}

// errorResponse is a JSON error response body.
type errorResponse struct {
	Error string `json:"error"`
}

// Handler provides HTTP handlers for the schema generation API.
type Handler struct {
	registry *Registry
}

// NewHandler creates a new Handler with the given type registry.
func NewHandler(registry *Registry) *Handler {
	return &Handler{registry: registry}
}

// GenerateHandler handles POST /schema/generate.
// It accepts a JSON body with either a "type" field (registered Go type)
// or a "data" field (raw JSON object) and returns both schemas.
func (h *Handler) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed, use POST")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close() //nolint:errcheck // closing body, error is irrelevant

	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, "request body is empty")
		return
	}

	var req generateRequest
	if unmarshalErr := json.Unmarshal(body, &req); unmarshalErr != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON in request body")
		return
	}

	var resp generateResponse

	switch {
	case req.Type != "":
		resp, err = h.generateFromType(req.Type)
	case len(req.Data) > 0:
		resp, err = h.generateFromData(req.Data)
	default:
		writeError(w, http.StatusBadRequest, "request must contain \"type\" or \"data\" field")
		return
	}

	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, parser.ErrInvalidJSON) || errors.Is(err, parser.ErrNotJSONObject) {
			status = http.StatusBadRequest
		}
		// Check for registry lookup errors (type not found).
		if isNotFoundError(err) {
			status = http.StatusNotFound
		}

		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// generateFromType generates schemas from a registered Go type name.
func (h *Handler) generateFromType(typeName string) (generateResponse, error) {
	v, err := h.registry.Lookup(typeName)
	if err != nil {
		return generateResponse{}, err
	}

	jsonSchema, err := parser.GenerateJSONSchema(v)
	if err != nil {
		return generateResponse{}, err
	}

	uiSchema, err := parser.GenerateUISchema(v)
	if err != nil {
		return generateResponse{}, err
	}

	return generateResponse{Schema: jsonSchema, UISchema: uiSchema}, nil
}

// generateFromData generates schemas from raw JSON data.
func (h *Handler) generateFromData(data json.RawMessage) (generateResponse, error) {
	jsonSchema, uiSchema, err := parser.GenerateFromJSON(data)
	if err != nil {
		return generateResponse{}, err
	}

	return generateResponse{Schema: jsonSchema, UISchema: uiSchema}, nil
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v) //nolint:errcheck // best-effort response write
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

// isNotFoundError checks if an error is a "not found" registry error.
func isNotFoundError(err error) bool {
	return err != nil && len(err.Error()) > 4 && err.Error()[len(err.Error())-len("not found in registry"):] == "not found in registry"
}
