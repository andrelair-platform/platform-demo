package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Build-time metadata. Overridden via -ldflags at build:
//   go build -ldflags "-X main.version=$(git rev-parse --short HEAD) -X main.commit=$(git rev-parse HEAD)"
var (
	version = "dev"
	commit  = "unknown"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, handler, and status code.",
	}, []string{"method", "handler", "code"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"method", "handler", "code"})
)

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.code = code
	sw.ResponseWriter.WriteHeader(code)
}

func instrument(pattern string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
		start := time.Now()
		h(sw, r)
		code := strconv.Itoa(sw.code)
		httpRequestsTotal.WithLabelValues(r.Method, pattern, code).Inc()
		httpRequestDuration.WithLabelValues(r.Method, pattern, code).Observe(time.Since(start).Seconds())
	}
}

type response struct {
	App       string `json:"app"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Hostname  string `json:"hostname"`
	GoVersion string `json:"goVersion"`
	Now       string `json:"now"`
	Message   string `json:"message"`
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response{
		App:       "platform-demo",
		Version:   version,
		Commit:    commit,
		Hostname:  hostname,
		GoVersion: runtime.Version(),
		Now:       time.Now().UTC().Format(time.RFC3339),
		Message:   "deployed end-to-end via GitHub Actions + ghcr.io + ArgoCD GitOps",
	})
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready\n"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9898"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", instrument("/", handleRoot))
	mux.HandleFunc("/healthz", instrument("/healthz", handleHealthz))
	mux.HandleFunc("/readyz", instrument("/readyz", handleReadyz))
	mux.Handle("/metrics", promhttp.Handler())
	addr := ":" + port
	log.Printf("platform-demo %s (%s) listening on %s", version, commit, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
