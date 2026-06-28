package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const appVersion = "multi-n-api-2"

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

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeHealth(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"service":      "api",
		"version":      appVersion,
		"git_sha":      buildSHA,
		"git_ref":      buildRef,
		"build_label":  buildLabel,
		"greeting_set": true,
	})
}

func probeService(name, baseURL, path string) map[string]any {
	out := map[string]any{
		"name":   name,
		"url":    strings.TrimRight(baseURL, "/") + path,
		"status": "unknown",
	}
	if strings.TrimSpace(baseURL) == "" {
		out["status"] = "skipped"
		out["reason"] = "không có URL discovery (SVC_" + strings.ToUpper(name) + "_URL)"
		return out
	}
	client := &http.Client{Timeout: 4 * time.Second}
	target := strings.TrimRight(baseURL, "/") + path
	resp, err := client.Get(target)
	if err != nil {
		out["status"] = "error"
		out["error"] = err.Error()
		return out
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	out["http_status"] = resp.StatusCode
	out["ok"] = resp.StatusCode >= 200 && resp.StatusCode < 300
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		out["status"] = "ok"
	} else {
		out["status"] = "degraded"
	}
	var parsed any
	if json.Unmarshal(body, &parsed) == nil {
		out["body"] = parsed
	} else {
		out["body"] = strings.TrimSpace(string(body))
	}
	return out
}

func fleetHandler(w http.ResponseWriter, _ *http.Request) {
	workerURL := strings.TrimSpace(os.Getenv("SVC_WORKER_URL"))
	webURL := strings.TrimSpace(os.Getenv("SVC_WEB_URL"))

	services := []map[string]any{
		{
			"name":    "api",
			"role":    "backend",
			"public":  true,
			"ingress": "/api",
			"status":  "ok",
			"version": appVersion,
			"git_ref": buildRef,
		},
		probeService("web", webURL, "/"),
		probeService("worker", workerURL, "/status"),
	}

	for i := range services {
		switch services[i]["name"] {
		case "web":
			services[i]["role"] = "frontend"
			services[i]["public"] = true
			services[i]["ingress"] = "/"
		case "worker":
			services[i]["role"] = "internal"
			services[i]["public"] = false
			services[i]["ingress"] = nil
			services[i]["discovery"] = "SVC_WORKER_URL"
		}
	}

	publicCount, internalCount := 0, 0
	for _, s := range services {
		pub, _ := s["public"].(bool)
		if pub {
			publicCount++
		} else {
			internalCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"layout":  "multi-n-service",
		"level":   "L3",
		"summary": fmt.Sprintf("%d services · %d public · %d internal", len(services), publicCount, internalCount),
		"fleet": map[string]any{
			"total":    len(services),
			"public":   publicCount,
			"internal": internalCount,
		},
		"services": services,
		"api": map[string]any{
			"version": appVersion,
			"git_ref": buildRef,
			"git_sha": buildSHA,
		},
		"note": "Worker không có Ingress — browser chỉ thấy qua API gọi nội bộ cluster (SVC_WORKER_URL).",
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
	http.HandleFunc("/api/fleet", fleetHandler)
	http.HandleFunc("/api/greeting", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		writeJSON(w, http.StatusOK, map[string]string{
			"greeting":    greeting,
			"build_label": buildLabel,
			"git_sha":     buildSHA,
			"git_ref":     buildRef,
			"version":     appVersion,
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
