//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// server is shared across all integration tests in this package.
var (
	integServer     *httptest.Server
	integServerOnce sync.Once
)

func integrationServer(t *testing.T) string {
	t.Helper()
	integServerOnce.Do(func() {
		integServer = httptest.NewServer(newHandler())
	})
	t.Cleanup(func() {}) // server lives for the whole test run
	return integServer.URL
}

func get(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(integrationServer(t) + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func TestIntegration_Root(t *testing.T) {
	resp := get(t, "/")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected application/json Content-Type, got %q", ct)
	}

	var body response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.App != "platform-demo" {
		t.Errorf("app=%q, want platform-demo", body.App)
	}
	if body.GoVersion == "" {
		t.Error("goVersion field is empty")
	}
	if body.Now == "" {
		t.Error("now field is empty")
	}
}

func TestIntegration_Healthz(t *testing.T) {
	resp := get(t, "/healthz")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.HasPrefix(string(body), "ok") {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestIntegration_Readyz(t *testing.T) {
	resp := get(t, "/readyz")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_Metrics(t *testing.T) {
	// Warm up the counters first
	get(t, "/healthz").Body.Close()
	get(t, "/readyz").Body.Close()

	resp := get(t, "/metrics")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain Content-Type for Prometheus, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "http_requests_total") {
		t.Error("metrics response missing http_requests_total")
	}
	if !strings.Contains(string(body), "http_request_duration_seconds") {
		t.Error("metrics response missing http_request_duration_seconds")
	}
}

func TestIntegration_Concurrent(t *testing.T) {
	const workers = 10
	const reqs = 5

	errs := make(chan error, workers*reqs)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < reqs; i++ {
				resp, err := http.Get(integrationServer(t) + "/healthz")
				if err != nil {
					errs <- fmt.Errorf("worker %d req %d: %w", id, i, err)
					return
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errs <- fmt.Errorf("worker %d req %d: status %d", id, i, resp.StatusCode)
				}
			}
		}(w)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
