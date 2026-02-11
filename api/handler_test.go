package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handler "github.com/holdemlab/ui-json-schema/api"
)

const (
	endpointPath    = "/schema/generate"
	contentTypeJSON = "application/json"
)

// newTestHandler creates a Handler with a pre-populated registry for tests.
func newTestHandler() *handler.Handler {
	reg := handler.NewRegistry()
	reg.Register("User", testUser{})

	return handler.NewHandler(reg)
}

// doPost is a test helper that sends a POST request and returns the recorder.
func doPost(t *testing.T, h *handler.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, endpointPath, strings.NewReader(body))
	req.Header.Set("Content-Type", contentTypeJSON)

	rr := httptest.NewRecorder()
	h.GenerateHandler(rr, req)

	return rr
}

// --- Tests: successful responses ---

func TestHandler_GenerateFromType(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"type":"User"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	// Check schema.
	s, ok := resp["schema"].(map[string]any)
	if !ok {
		t.Fatal("expected 'schema' object in response")
	}

	if s["type"] != "object" {
		t.Errorf("expected schema type 'object', got %v", s["type"])
	}

	props, ok := s["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected 'properties' in schema")
	}

	if _, ok := props["name"]; !ok {
		t.Error("expected 'name' property in schema")
	}

	// Check required.
	req, ok := s["required"].([]any)
	if !ok {
		t.Fatal("expected 'required' array")
	}

	if len(req) != 1 || req[0] != "name" {
		t.Errorf("expected required=[name], got %v", req)
	}

	// Check uischema.
	ui, ok := resp["uischema"].(map[string]any)
	if !ok {
		t.Fatal("expected 'uischema' object in response")
	}

	if ui["type"] != "VerticalLayout" {
		t.Errorf("expected VerticalLayout, got %v", ui["type"])
	}
}

func TestHandler_GenerateFromData(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"data":{"name":"John","age":30,"active":true}}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	s, ok := resp["schema"].(map[string]any)
	if !ok {
		t.Fatal("expected 'schema' object")
	}

	props, ok := s["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected 'properties' in schema")
	}

	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}

	// From JSON, all fields should be optional (no required).
	if _, hasRequired := s["required"]; hasRequired {
		t.Error("expected no 'required' for JSON-generated schema")
	}

	ui, ok := resp["uischema"].(map[string]any)
	if !ok {
		t.Fatal("expected 'uischema' object")
	}

	if ui["type"] != "VerticalLayout" {
		t.Errorf("expected VerticalLayout, got %v", ui["type"])
	}
}

func TestHandler_GenerateFromData_NestedObject(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"data":{"address":{"city":"Kyiv","zip":"01001"}}}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	s := resp["schema"].(map[string]any)
	props := s["properties"].(map[string]any)
	addr := props["address"].(map[string]any)

	if addr["type"] != "object" {
		t.Errorf("expected nested object type, got %v", addr["type"])
	}
}

func TestHandler_TypeTakesPriority(t *testing.T) {
	h := newTestHandler()
	// When both type and data are present, type wins.
	rr := doPost(t, h, `{"type":"User","data":{"x":1}}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	s := resp["schema"].(map[string]any)
	props := s["properties"].(map[string]any)

	// Should have User's fields, not "x".
	if _, ok := props["name"]; !ok {
		t.Error("expected 'name' property from User type")
	}
	if _, ok := props["x"]; ok {
		t.Error("unexpected 'x' property — type should take priority over data")
	}
}

func TestHandler_ContentType(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"type":"User"}`)

	ct := rr.Header().Get("Content-Type")
	if ct != contentTypeJSON {
		t.Errorf("expected Content-Type %q, got %q", contentTypeJSON, ct)
	}
}

// --- Tests: error responses ---

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, endpointPath, nil)
	rr := httptest.NewRecorder()
	h.GenerateHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}

	assertErrorResponse(t, rr)
}

func TestHandler_EmptyBody(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, "")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}

	assertErrorResponse(t, rr)
}

func TestHandler_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{invalid}`)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}

	assertErrorResponse(t, rr)
}

func TestHandler_NoTypeOrData(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{}`)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	assertErrorResponse(t, rr)
}

func TestHandler_TypeNotFound(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"type":"NonExistent"}`)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}

	assertErrorResponse(t, rr)
}

func TestHandler_InvalidDataPayload(t *testing.T) {
	h := newTestHandler()
	// data is a JSON array, not an object.
	rr := doPost(t, h, `{"data":[1,2,3]}`)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	assertErrorResponse(t, rr)
}

func TestHandler_InvalidDataString(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"data":"not an object"}`)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	assertErrorResponse(t, rr)
}

func TestHandler_PUTMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPut, endpointPath, nil)
	rr := httptest.NewRecorder()
	h.GenerateHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandler_DELETEMethod(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodDelete, endpointPath, nil)
	rr := httptest.NewRecorder()
	h.GenerateHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// --- Tests: response structure validation ---

func TestHandler_ResponseStructure_FromType(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"type":"User"}`)

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Must have exactly "schema" and "uischema" top-level keys.
	if _, ok := resp["schema"]; !ok {
		t.Error("missing 'schema' in response")
	}
	if _, ok := resp["uischema"]; !ok {
		t.Error("missing 'uischema' in response")
	}
}

func TestHandler_ResponseStructure_FromData(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"data":{"x":1}}`)

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := resp["schema"]; !ok {
		t.Error("missing 'schema' in response")
	}
	if _, ok := resp["uischema"]; !ok {
		t.Error("missing 'uischema' in response")
	}
}

func TestHandler_SchemaHasDraftField(t *testing.T) {
	h := newTestHandler()
	rr := doPost(t, h, `{"type":"User"}`)

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	s := resp["schema"].(map[string]any)
	draft, ok := s["$schema"]
	if !ok {
		t.Error("expected $schema field")
	}
	if draft != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("unexpected $schema value: %v", draft)
	}
}

// --- helpers ---

func assertErrorResponse(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected JSON error response, got: %s", rr.Body.String())
	}

	if _, ok := resp["error"]; !ok {
		t.Error("expected 'error' field in error response")
	}
}
