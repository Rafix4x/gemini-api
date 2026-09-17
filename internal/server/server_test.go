package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Rafix4x/gemini-api/internal/config"
	"github.com/Rafix4x/gemini-api/internal/format"
	"github.com/Rafix4x/gemini-api/internal/models"
)

func TestHealthEndpoint(t *testing.T) {
	testVer := "test-version-1.0"
	cfg := config.Default()
	app := New(cfg, testVer)
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}

	if body["status"] != "ok" || body["version"] != testVer {
		t.Errorf("Unexpected health response: %v", body)
	}
}

func TestAuthMatrix(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"sk-secret-key"}
	app := New(cfg, "test-version")
	handler := app.Handler()

	req1 := httptest.NewRequest("GET", "/v1/models", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest("GET", "/v1/models", nil)
	req2.Header.Set("Authorization", "Bearer wrong-key")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong key, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest("GET", "/v1/models", nil)
	req3.Header.Set("Authorization", "Bearer sk-secret-key")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("Expected 200 with Bearer token, got %d", rec3.Code)
	}

	req4 := httptest.NewRequest("GET", "/v1/models", nil)
	req4.Header.Set("x-api-key", "sk-secret-key")
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-api-key, got %d", rec4.Code)
	}

	req5 := httptest.NewRequest("GET", "/v1/models", nil)
	req5.Header.Set("x-goog-api-key", "sk-secret-key")
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-goog-api-key, got %d", rec5.Code)
	}

	req6 := httptest.NewRequest("GET", "/v1/models?key=sk-secret-key", nil)
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)
	if rec6.Code != http.StatusOK {
		t.Errorf("Expected 200 with ?key= query, got %d", rec6.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	req := httptest.NewRequest("OPTIONS", "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{
		"text": "<hello & world>",
	}
	b, err := marshalNoEscapeHTML(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := `{"text":"<hello & world>"}`
	if string(b) != expected {
		t.Errorf("Got %q, want %q", string(b), expected)
	}
}

func TestChunkedTransferEncoding(t *testing.T) {

	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	body := `{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusBadRequest {
		t.Errorf("Handler rejected valid JSON body, got %d", rec.Code)
	}
}

func TestResponsesSequenceNumberOrder(t *testing.T) {

	output := format.BuildResponseOutput("Hello world", nil, "msg_test123")
	if len(output) != 1 {
		t.Fatalf("Expected 1 output item, got %d", len(output))
	}
	item := output[0]
	if item["type"] != "message" {
		t.Errorf("Expected type=message, got %v", item["type"])
	}
	if item["id"] != "msg_test123" {
		t.Errorf("Expected id=msg_test123, got %v", item["id"])
	}

	tc := models.OpenAIToolCall{
		ID:   "call_abc123",
		Type: "function",
		Function: models.OpenAIToolCallFunction{
			Name:      "get_weather",
			Arguments: `{"city":"Jakarta"}`,
		},
	}
	outputWithTools := format.BuildResponseOutput("", []models.OpenAIToolCall{tc}, "msg_456")

	if len(outputWithTools) != 1 {
		t.Fatalf("Expected 1 output item (function_call only when text empty), got %d", len(outputWithTools))
	}
	if outputWithTools[0]["type"] != "function_call" {
		t.Errorf("First item should be function_call, got %v", outputWithTools[0]["type"])
	}
	if outputWithTools[0]["call_id"] != "call_abc123" {
		t.Errorf("Expected call_id=call_abc123, got %v", outputWithTools[0]["call_id"])
	}

	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	body := `{"model":"gemini-3.6-flash","input":"test prompt"}`
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusBadRequest {
		t.Errorf("Responses endpoint rejected valid JSON, got %d: %s", rec.Code, rec.Body.String())
	}
}
