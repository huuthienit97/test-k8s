package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const appVersion = "auto-deploy-test-2"

func main() {
	log.SetFlags(log.LstdFlags)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		log.Printf("health check")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","version":"` + appVersion + `"}`))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "test-k8s v%s says hello! path=%s\n", appVersion, r.URL.Path)
	})

	log.Printf("server listening on :%s (version %s)", port, appVersion)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
