package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func redisConfigured() bool {
	return strings.TrimSpace(os.Getenv("REDIS_URL")) != ""
}

func redisClient() *redis.Client {
	u := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if u == "" {
		return nil
	}
	opt, err := redis.ParseURL(u)
	if err != nil {
		return nil
	}
	opt.DialTimeout = 4 * time.Second
	opt.ReadTimeout = 4 * time.Second
	opt.WriteTimeout = 4 * time.Second
	return redis.NewClient(opt)
}

func redisStatus(ctx context.Context) map[string]any {
	out := map[string]any{
		"name":       "redis",
		"role":       "data",
		"stack":      "redis",
		"public":     false,
		"configured": redisConfigured(),
	}
	if !redisConfigured() {
		out["status"] = "skipped"
		out["reason"] = "REDIS_URL chưa inject — bật addon Redis trên Console"
		return out
	}
	c := redisClient()
	if c == nil {
		out["status"] = "error"
		out["error"] = "REDIS_URL không parse được"
		return out
	}
	defer c.Close()
	start := time.Now()
	if err := c.Ping(ctx).Err(); err != nil {
		out["status"] = "error"
		out["error"] = err.Error()
		return out
	}
	out["status"] = "ok"
	out["ok"] = true
	out["latency_ms"] = time.Since(start).Milliseconds()
	out["discovery"] = "REDIS_URL"
	val, err := c.Get(ctx, "console:hello").Result()
	if err == nil {
		out["demo_key"] = "console:hello"
		out["demo_value"] = val
	}
	return out
}

func redisPingHandler(w http.ResponseWriter, r *http.Request) {
	c := redisClient()
	if c == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"redis":      "unconfigured",
			"configured": false,
			"hint":       "Bật Redis addon trên Platform Console",
		})
		return
	}
	defer c.Close()
	start := time.Now()
	if err := c.Ping(r.Context()).Err(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"redis": "error", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"redis":       "pong",
		"configured":  true,
		"latency_ms":  time.Since(start).Milliseconds(),
		"service":     "api",
		"via":         "REDIS_URL",
	})
}

func redisSetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}
	c := redisClient()
	if c == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "REDIS_URL chưa cấu hình"})
		return
	}
	defer c.Close()
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	key := strings.TrimSpace(body.Key)
	if key == "" {
		key = "console:hello"
	}
	val := body.Value
	if val == "" {
		val = "world-from-ui"
	}
	ctx := r.Context()
	if err := c.Set(ctx, key, val, time.Hour).Err(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"action": "set", "key": key, "value": val, "ttl": "1h"})
}

func redisGetHandler(w http.ResponseWriter, r *http.Request) {
	c := redisClient()
	if c == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "REDIS_URL chưa cấu hình"})
		return
	}
	defer c.Close()
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		key = "console:hello"
	}
	val, err := c.Get(r.Context(), key).Result()
	if err == redis.Nil {
		writeJSON(w, http.StatusOK, map[string]any{"key": key, "value": nil, "result": "(nil)"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"key": key, "value": val, "result": val})
}

func redisDemoHandler(w http.ResponseWriter, r *http.Request) {
	c := redisClient()
	if c == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "REDIS_URL chưa cấu hình"})
		return
	}
	defer c.Close()
	ctx := r.Context()
	key := "console:hello"
	val := "world-from-api-" + appVersion
	if err := c.Set(ctx, key, val, time.Hour).Err(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	got, err := c.Get(ctx, key).Result()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"action": "set+get",
		"key":    key,
		"value":  got,
		"ttl":    "1h",
	})
}
