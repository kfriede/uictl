package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Error("expected API key header")
		}
		if r.Header.Get("User-Agent") != userAgent {
			t.Errorf("expected User-Agent=%s, got %s", userAgent, r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"applicationVersion": "10.2.93"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	// Override baseURL since test server doesn't have /integration prefix
	c.baseURL = server.URL
	c.pathDetected = true

	var result map[string]string
	if err := c.GetJSON("", &result); err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if result["applicationVersion"] != "10.2.93" {
		t.Errorf("expected version 10.2.93, got %s", result["applicationVersion"])
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	c.baseURL = server.URL
	c.pathDetected = true

	var result map[string]string
	if err := c.GetJSON("", &result); err != nil {
		t.Fatalf("GetJSON failed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"statusCode": 404,
			"statusName": "NOT_FOUND",
			"message":    "Device not found",
		})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	c.baseURL = server.URL
	c.pathDetected = true

	_, err := c.Get("")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestClientPagination(t *testing.T) {
	page := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := PageResponse{
			Limit:      2,
			TotalCount: 5,
		}
		switch page {
		case 1:
			resp.Offset = 0
			resp.Count = 2
			resp.Data = []map[string]any{{"id": "1"}, {"id": "2"}}
		case 2:
			resp.Offset = 2
			resp.Count = 2
			resp.Data = []map[string]any{{"id": "3"}, {"id": "4"}}
		case 3:
			resp.Offset = 4
			resp.Count = 1
			resp.Data = []map[string]any{{"id": "5"}}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	c.baseURL = server.URL
	c.pathDetected = true

	results, err := c.GetAllPages("")
	if err != nil {
		t.Fatalf("GetAllPages failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 total results, got %d", len(results))
	}
}
