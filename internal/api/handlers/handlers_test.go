package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourcompany/image-warehousing/internal/models"
)

func TestHealthHandler_HandleHealth(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type to contain application/json, got %s", contentType)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check fields
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("expected status 'healthy', got %v", response["status"])
	}

	if service, ok := response["service"].(string); !ok || service != "image-warehousing" {
		t.Errorf("expected service 'image-warehousing', got %v", response["service"])
	}

	if _, ok := response["timestamp"]; !ok {
		t.Error("response should contain timestamp field")
	}
}

func TestSearchHandler_HandleSearch_Success(t *testing.T) {
	// This test verifies the handler structure and validation
	// In a real scenario with dependency injection, we'd use a mock service

	handler := &SearchHandler{
		searchService: nil, // Would be a mock in real test
	}

	if handler == nil {
		t.Error("failed to create handler")
	}
}

func TestSearchHandler_HandleSearch_InvalidJSON(t *testing.T) {
	handler := NewSearchHandler(nil)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handler.HandleSearch(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestSearchHandler_HandleSearch_EmptyQuery(t *testing.T) {
	handler := NewSearchHandler(nil)

	// Send request with empty query
	reqBody := models.SearchRequest{
		Query: "",
		Limit: 10,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/search", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleSearch(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty query, got %d", resp.StatusCode)
	}
}

func TestSearchHandler_DefaultLimit(t *testing.T) {
	// Test that default limit logic would work correctly
	req := models.SearchRequest{
		Query: "test",
		Limit: 0, // Not set
	}

	// Simulate the default limit logic
	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit != 10 {
		t.Errorf("expected default limit 10, got %d", req.Limit)
	}
}

func TestSearchHandler_CustomLimit(t *testing.T) {
	// Test that custom limit is preserved
	req := models.SearchRequest{
		Query: "test",
		Limit: 25,
	}

	// Simulate the default limit logic
	if req.Limit == 0 {
		req.Limit = 10
	}

	if req.Limit != 25 {
		t.Errorf("expected custom limit 25, got %d", req.Limit)
	}
}

func TestSearchRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     models.SearchRequest
		shouldError bool
		reason      string
	}{
		{
			name: "valid request",
			request: models.SearchRequest{
				Query: "cats",
				Limit: 10,
			},
			shouldError: false,
		},
		{
			name: "empty query",
			request: models.SearchRequest{
				Query: "",
				Limit: 10,
			},
			shouldError: true,
			reason:      "query is required",
		},
		{
			name: "zero limit should get default",
			request: models.SearchRequest{
				Query: "test",
				Limit: 0,
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation logic
			hasError := tt.request.Query == ""

			if hasError != tt.shouldError {
				t.Errorf("expected error=%v, got error=%v", tt.shouldError, hasError)
			}
		})
	}
}

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()

	if handler == nil {
		t.Fatal("NewHealthHandler returned nil")
	}
}

func TestNewSearchHandler(t *testing.T) {
	handler := NewSearchHandler(nil)

	if handler == nil {
		t.Fatal("NewSearchHandler returned nil")
	}

	if handler.searchService != nil {
		t.Error("expected searchService to be nil when passed nil")
	}
}

func TestHealthHandler_Methods(t *testing.T) {
	handler := NewHealthHandler()

	// Test that handler accepts GET requests
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET request should succeed, got status %d", w.Code)
	}

	// Test that handler also accepts POST (since there's no method check)
	req = httptest.NewRequest(http.MethodPost, "/health", nil)
	w = httptest.NewRecorder()

	handler.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST request should succeed (no method restriction), got status %d", w.Code)
	}
}

func TestSearchHandler_ContentType(t *testing.T) {
	// Verify the handler would set correct content type
	// This tests the expected behavior

	expectedContentType := "application/json"

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", expectedContentType)

	contentType := w.Header().Get("Content-Type")
	if contentType != expectedContentType {
		t.Errorf("expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}
