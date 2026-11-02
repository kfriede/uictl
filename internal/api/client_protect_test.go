package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProtectAndNetworkRouting(t *testing.T) {
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/proxy/protect/integration/v1/alarm-hubs":
			_, _ = w.Write([]byte("[]"))
		case "/proxy/network/integration/v1/info":
			_ = json.NewEncoder(w).Encode(map[string]string{"applicationVersion": "10.0.0"})
		case "/proxy/network/integration/v1/sites":
			_, _ = w.Write([]byte(`{"offset":0,"limit":25,"count":0,"totalCount":0,"data":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})

	if _, err := c.Get("/v1/alarm-hubs"); err != nil {
		t.Fatalf("protect get: %v", err)
	}
	if len(paths) != 1 || paths[0] != "/proxy/protect/integration/v1/alarm-hubs" {
		t.Fatalf("protect path = %v, want [/proxy/protect/integration/v1/alarm-hubs]", paths)
	}

	if _, err := c.Get("/v1/sites"); err != nil {
		t.Fatalf("network get: %v", err)
	}
	sawSites := false
	for _, p := range paths[1:] {
		if strings.HasPrefix(p, "/proxy/protect/") {
			t.Fatalf("network request used protect base: %v", paths)
		}
		if p == "/proxy/network/integration/v1/sites" {
			sawSites = true
		}
	}
	if !sawSites {
		t.Fatalf("network path = %v, want detected /proxy/network/integration/v1/sites", paths)
	}
}

func TestProtectBaseNotDoubled(t *testing.T) {
	var got string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
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
	if _, err := c.Get("/v1/cameras"); err != nil {
		t.Fatal(err)
	}
	if got != "/proxy/protect/integration/v1/cameras" {
		t.Fatalf("path = %s", got)
	}
}

func TestPostFileProtectPath(t *testing.T) {
	var gotPath, gotCT string
	var gotBody []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"file-1","type":"ringtone","path":"/data/file-1"}`))
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	data, err := c.PostFile("/v1/files/ringtone", "tone.mp3", strings.NewReader("audio"))
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/proxy/protect/integration/v1/files/ringtone" {
		t.Fatalf("path = %s", gotPath)
	}
	if !strings.HasPrefix(gotCT, "multipart/form-data;") {
		t.Fatalf("content-type = %s", gotCT)
	}
	if !strings.Contains(string(gotBody), `name="file"`) || !strings.Contains(string(gotBody), "tone.mp3") {
		t.Fatalf("multipart body = %s", gotBody)
	}
	if !strings.Contains(string(data), "file-1") {
		t.Fatalf("response = %s", data)
	}
}

func TestSubscribeTextFrame(t *testing.T) {
	payload := []byte(`{"type":"deviceUpdate"}`)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxy/protect/integration/v1/subscribe/devices" {
			http.Error(w, "bad path "+r.URL.Path, http.StatusNotFound)
			return
		}
		if r.Header.Get("Upgrade") != "websocket" || r.Header.Get("X-API-KEY") != "test-key" {
			http.Error(w, "missing websocket headers", http.StatusBadRequest)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "no hijack", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		_ = bufrw.Flush()
		frame := append([]byte{0x81, byte(len(payload))}, payload...)
		_, _ = conn.Write(frame)
		_, _ = conn.Write([]byte{0x88, 0x00})
	}))
	defer server.Close()

	c := NewClient(ClientConfig{
		Host:      server.URL,
		APIKey:    "test-key",
		Insecure:  true,
		ErrWriter: os.Stderr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var got []byte
	err := c.Subscribe(ctx, "/v1/subscribe/devices", func(msg []byte) error {
		got = append([]byte{}, msg...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("message = %q", got)
	}
}
