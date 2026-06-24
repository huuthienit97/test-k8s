package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const appVersion = "env-demo-2"

// App đọc APP_GREETING từ môi trường (K8s Secret app-env trên Platform Console).
// Không có file .env trong Git — khai báo trên tab Env vars.
func requiredEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf(
			"thiếu biến môi trường %s — thêm trên Platform Console → Project → Env vars (dev/prod), không commit .env lên Git",
			key,
		)
	}
	return v
}

func main() {
	log.SetFlags(log.LstdFlags)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	greeting := requiredEnv("APP_GREETING")

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","version":"` + appVersion + `","greeting_set":true}`))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(
			w,
			"test-k8s v%s\nAPP_GREETING=%s\npath=%s\n",
			appVersion, greeting, r.URL.Path,
		)
	})

	log.Printf("server listening on :%s (version %s, APP_GREETING ok)", port, appVersion)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
