package tools

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Call the Get function with the test server URL
	body, err := Get(server.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that the response body is correct
	expectedBody := "test response"
	actualBody, err := io.ReadAll(body)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(actualBody) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(actualBody))
	}
}
