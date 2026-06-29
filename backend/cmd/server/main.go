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

const appVersion = "polyglot-submodule-api-3"

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
		"stack":        "go",
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
		out["reason"] = "không có SVC_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_URL"
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

func internalServices() []struct {
	name, envKey, path string
	role               string
	stack              string
} {
	return []struct {
		name, envKey, path string
		role               string
		stack              string
	}{
		{"node", "SVC_NODE_URL", "/hello", "backend", "node"},
		{"dotnet", "SVC_DOTNET_URL", "/hello", "backend", "dotnet"},
		{"worker", "SVC_WORKER_URL", "/status", "internal", "python"},
	}
}

func polyglotHandler(w http.ResponseWriter, _ *http.Request) {
	out := map[string]any{
		"gateway": map[string]any{
			"service": "api",
			"stack":   "go",
			"version": appVersion,
			"git_ref": buildRef,
		},
		"level":  "L4C",
		"layout": "multi-submodules",
		"git": map[string]any{
			"submodules": "recursive",
			"lib":        "libs/is-docker",
		},
	}
	backends := make([]map[string]any, 0, 3)
	for _, s := range internalServices() {
		item := probeService(s.name, os.Getenv(s.envKey), s.path)
		item["stack"] = s.stack
		item["role"] = s.role
		backends = append(backends, item)
	}
	out["backends"] = backends
	writeJSON(w, http.StatusOK, out)
}

func fleetHandler(w http.ResponseWriter, _ *http.Request) {
	services := []map[string]any{
		{
			"name": "api", "role": "gateway", "stack": "go", "public": true,
			"ingress": "/api", "status": "ok", "version": appVersion, "git_ref": buildRef,
		},
	}
	webProbe := probeService("web", os.Getenv("SVC_WEB_URL"), "/")
	webProbe["role"] = "frontend"
	webProbe["stack"] = "react"
	webProbe["public"] = true
	webProbe["ingress"] = "/"
	services = append(services, webProbe)

	for _, s := range internalServices() {
		item := probeService(s.name, os.Getenv(s.envKey), s.path)
		item["role"] = s.role
		item["stack"] = s.stack
		item["public"] = false
		item["discovery"] = s.envKey
		services = append(services, item)
	}

	publicCount, internalCount := 0, 0
	for _, s := range services {
		if pub, _ := s["public"].(bool); pub {
			publicCount++
		} else {
			internalCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"layout":  "multi-submodules",
		"level":   "L4C",
		"summary": fmt.Sprintf("%d services · %d public · %d internal · auto-deploy demo v3", len(services), publicCount, internalCount),
		"fleet": map[string]any{
			"total": len(services), "public": publicCount, "internal": internalCount,
		},
		"services": services,
		"stacks":   []string{"react", "go", "node", "dotnet", "python"},
		"api": map[string]any{
			"version": appVersion, "git_ref": buildRef, "git_sha": buildSHA,
		},
	})
}

func callBackendHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/call/"), "/")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu tên service — vd. /api/call/node"})
		return
	}
	if name == "go" || name == "api" {
		writeJSON(w, http.StatusOK, map[string]any{
			"name": "api", "stack": "go", "via": "go-gateway", "status": "ok",
			"version": appVersion, "git_ref": buildRef, "git_sha": buildSHA,
			"url": "/api/health",
		})
		return
	}
	for _, s := range internalServices() {
		if s.name != name {
			continue
		}
		result := probeService(s.name, os.Getenv(s.envKey), s.path)
		result["stack"] = s.stack
		result["role"] = s.role
		result["via"] = "go-gateway"
		result["discovery"] = s.envKey
		writeJSON(w, http.StatusOK, result)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{
		"error": "service không tồn tại: " + name,
		"hint":  "dùng node | dotnet | worker | go",
	})
}

func main() {
	log.SetFlags(log.LstdFlags)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	greeting := requiredEnv("APP_GREETING")

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { writeHealth(w) })
	http.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) { writeHealth(w) })
	http.HandleFunc("/api/fleet", fleetHandler)
	http.HandleFunc("/api/polyglot", polyglotHandler)
	http.HandleFunc("/api/call/", callBackendHandler)
	http.HandleFunc("/api/greeting", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		writeJSON(w, http.StatusOK, map[string]string{
			"greeting": greeting, "build_label": buildLabel,
			"git_sha": buildSHA, "git_ref": buildRef, "version": appVersion,
		})
	})

	log.Printf("api (go) listening on :%s (%s)", port, appVersion)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
