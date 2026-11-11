package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestProtectRouteExternalArmProfiles(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	if _, err := c.Get("/v1/external/arm-profiles"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	want := "/proxy/protect/integration/v1/external/arm-profiles"
	if len(paths) != 1 || paths[0] != want {
		t.Errorf("paths = %v, want [%s]", paths, want)
	}
	for _, p := range paths {
		if p == "/proxy/network/integration/v1/info" || strings.Contains(p, "/proxy/network/") {
			t.Errorf("Protect request ran Network detection: %v", paths)
		}
	}
}

func TestProtectRouteSubscribeUpdatesKeepsQuery(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	if _, err := c.Get("/v1/subscribe/updates?modelKeys=camera"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	want := "/proxy/protect/integration/v1/subscribe/updates?modelKeys=camera"
	if len(paths) != 1 || paths[0] != want {
		t.Errorf("paths = %v, want [%s]", paths, want)
	}
	for _, p := range paths {
		if strings.Contains(p, "/proxy/network/") || p == "/proxy/network/integration/v1/info" {
			t.Errorf("Protect subscribe ran Network detection: %v", paths)
		}
	}
}

func TestNetworkSitesSkipProtectBase(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	if _, err := c.Get("/v1/sites"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("expected Network requests")
	}
	usedNetwork := false
	for _, p := range paths {
		if strings.Contains(p, "/proxy/protect/") {
			t.Errorf("Network path used Protect base: %v", paths)
		}
		if strings.HasPrefix(p, "/proxy/network/integration") || strings.HasPrefix(p, "/integration") {
			usedNetwork = true
		}
	}
	if !usedNetwork {
		t.Errorf("expected Network detection path, got %v", paths)
	}
}

func TestProtectBaseNotDoubled(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL + "/proxy/protect/integration",
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	if _, err := c.Get("/v1/external/arm-profiles"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	want := "/proxy/protect/integration/v1/external/arm-profiles"
	doubled := "/proxy/protect/integration/proxy/protect/integration/v1/external/arm-profiles"
	if len(paths) != 1 || paths[0] != want {
		t.Errorf("paths = %v, want [%s]", paths, want)
	}
	for _, p := range paths {
		if p == doubled || strings.Count(p, "/proxy/protect/integration") > 1 {
			t.Errorf("Protect prefix was doubled: %v", paths)
		}
	}
}
