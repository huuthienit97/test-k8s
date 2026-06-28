package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const appVersion = "multi-api-3"

var buildSHA = "local"
var buildRef = "dev"
var buildLabel = ""

func requiredEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("thiếu biến môi trường %s — thêm trên Console → Env vars", key)
	}
	return v
}

func writeHealth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "ok",
		"service":      "api",
		"version":      appVersion,
		"git_sha":      buildSHA,
		"git_ref":      buildRef,
		"build_label":  buildLabel,
		"greeting_set": true,
	})
}

func main() {
	log.SetFlags(log.LstdFlags)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	greeting := requiredEnv("APP_GREETING")

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeHealth(w)
	})

	http.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeHealth(w)
	})

	http.HandleFunc("/api/greeting", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"greeting":     greeting,
			"build_label":  buildLabel,
			"git_sha":      buildSHA,
			"git_ref":      buildRef,
			"version":      appVersion,
		})
	})

	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "test-k8s api v%s\npath=%s\n", appVersion, r.URL.Path)
	})

	log.Printf("api listening on :%s (%s)", port, appVersion)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
