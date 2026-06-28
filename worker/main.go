package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const appVersion = "multi-n-worker-1"

type status struct {
	mu          sync.RWMutex
	APIURL      string `json:"api_url"`
	LastCheck   string `json:"last_check,omitempty"`
	LastOK      bool   `json:"last_ok"`
	LastStatus  int    `json:"last_status,omitempty"`
	LastError   string `json:"last_error,omitempty"`
	LastBody    string `json:"last_body,omitempty"`
}

func (s *status) snapshot() status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s
}

func (s *status) record(apiURL string, ok bool, code int, body, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.APIURL = apiURL
	s.LastCheck = time.Now().UTC().Format(time.RFC3339)
	s.LastOK = ok
	s.LastStatus = code
	s.LastBody = body
	s.LastError = errMsg
}

func main() {
	log.SetFlags(log.LstdFlags)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	apiURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SVC_API_URL")), "/")
	if apiURL == "" {
		log.Fatal("thiếu SVC_API_URL — platform inject service discovery khi deploy multi-service")
	}
	log.Printf("worker %s — ping API qua %s", appVersion, apiURL)

	st := &status{APIURL: apiURL}
	ping := func() {
		target := apiURL + "/api/health"
		resp, err := http.Get(target)
		if err != nil {
			st.record(apiURL, false, 0, "", err.Error())
			log.Printf("ping fail %s: %v", target, err)
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		ok := resp.StatusCode >= 200 && resp.StatusCode < 300
		st.record(apiURL, ok, resp.StatusCode, strings.TrimSpace(string(b)), "")
		log.Printf("ping %s → %d ok=%v", target, resp.StatusCode, ok)
	}
	ping()
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			ping()
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": "worker",
			"version": appVersion,
		})
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		snap := st.snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": "worker",
			"version": appVersion,
			"api_url": snap.APIURL,
			"check":   snap,
		})
	})

	log.Printf("worker listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
