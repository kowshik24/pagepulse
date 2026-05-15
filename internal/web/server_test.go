package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pagepulse/internal/buildinfo"
	"pagepulse/internal/metrics"
)

func TestServerSummaryEndpointAndHeaders(t *testing.T) {
	collector, err := metrics.NewCollector(time.Second)
	if err != nil {
		t.Fatalf("collector init failed: %v", err)
	}

	srv, err := NewServer(collector, buildinfo.Current())
	if err != nil {
		t.Fatalf("server init failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing X-Content-Type-Options header")
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content-type, got %q", rr.Header().Get("Content-Type"))
	}

	var summary metrics.Summary
	if err := json.Unmarshal(rr.Body.Bytes(), &summary); err != nil {
		t.Fatalf("summary response is not valid JSON: %v", err)
	}
	if summary.Hostname == "" {
		t.Fatal("expected hostname in summary")
	}
}

func TestServerSummaryRejectsMethod(t *testing.T) {
	collector, err := metrics.NewCollector(time.Second)
	if err != nil {
		t.Fatalf("collector init failed: %v", err)
	}

	srv, err := NewServer(collector, buildinfo.Current())
	if err != nil {
		t.Fatalf("server init failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/summary", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestServerVersionEndpoint(t *testing.T) {
	collector, err := metrics.NewCollector(time.Second)
	if err != nil {
		t.Fatalf("collector init failed: %v", err)
	}

	info := buildinfo.Info{
		Version:   "v1.2.3",
		Commit:    "abc123",
		BuildTime: "2026-05-16T00:00:00Z",
		GoVersion: "go1.test",
	}
	srv, err := NewServer(collector, info)
	if err != nil {
		t.Fatalf("server init failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got buildinfo.Info
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid version json: %v", err)
	}
	if got.Version != info.Version || got.Commit != info.Commit || got.BuildTime != info.BuildTime {
		t.Fatalf("unexpected version payload: %#v", got)
	}
}
