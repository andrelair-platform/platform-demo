package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Build-time metadata. Overridden via -ldflags at build:
//   go build -ldflags "-X main.version=$(git rev-parse --short HEAD) -X main.commit=$(git rev-parse HEAD)"
var (
	version = "dev"
	commit  = "unknown"
)

type response struct {
	App       string `json:"app"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Hostname  string `json:"hostname"`
	GoVersion string `json:"goVersion"`
	Now       string `json:"now"`
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
		// Match the bootstrap podinfo container port so the Service +
		// Ingress in minicloud-gitops keep working without a port edit.
		port = "9898"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", handleReadyz)
	addr := ":" + port
	log.Printf("platform-demo %s (%s) listening on %s", version, commit, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
