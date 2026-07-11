package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const appVersion = "multi-api-files-1"

var buildSHA = "local"
var buildRef = "dev"
var buildLabel = ""

type s3Config struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	Bucket         string
	Region         string
	UseSSL         bool
	ForcePathStyle bool
}

func requiredEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("thiếu biến môi trường %s — thêm trên Console → Env vars", key)
	}
	return v
}

func envOr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func loadS3Config() (s3Config, bool) {
	cfg := s3Config{
		Endpoint:       strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		AccessKey:      strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")),
		SecretKey:      strings.TrimSpace(os.Getenv("S3_SECRET_KEY")),
		Bucket:         envOr("S3_BUCKET", "app"),
		Region:         envOr("S3_REGION", "us-east-1"),
		UseSSL:         strings.EqualFold(os.Getenv("S3_USE_SSL"), "true"),
		ForcePathStyle: !strings.EqualFold(os.Getenv("S3_FORCE_PATH_STYLE"), "false"),
	}
	ok := cfg.Endpoint != "" && cfg.AccessKey != "" && cfg.SecretKey != "" && cfg.Bucket != ""
	return cfg, ok
}

func newS3Client(cfg s3Config) *s3.Client {
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	return s3.New(s3.Options{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, "",
		),
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: cfg.ForcePathStyle,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeHealth(w http.ResponseWriter, s3OK bool) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"service":      "api",
		"version":      appVersion,
		"git_sha":      buildSHA,
		"git_ref":      buildRef,
		"build_label":  buildLabel,
		"greeting_set": true,
		"s3_ready":     s3OK,
	})
}

func sanitizeObjectKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return "", fmt.Errorf("thiếu key")
	}
	if strings.Contains(key, "..") {
		return "", fmt.Errorf("key không hợp lệ")
	}
	return key, nil
}

func main() {
	log.SetFlags(log.LstdFlags)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	greeting := requiredEnv("APP_GREETING")
	s3cfg, s3OK := loadS3Config()
	var s3c *s3.Client
	if s3OK {
		s3c = newS3Client(s3cfg)
		log.Printf("S3 ready endpoint=%s bucket=%s", s3cfg.Endpoint, s3cfg.Bucket)
	} else {
		log.Printf("S3 chưa cấu hình — bật MinIO addon để inject S3_*")
	}

	mux := http.NewServeMux()

	writeOKHealth := func(w http.ResponseWriter, _ *http.Request) {
		writeHealth(w, s3OK)
	}
	// Probe platform: PublicHealthPath("/api","/health") → /api/health
	mux.HandleFunc("/health", writeOKHealth)
	mux.HandleFunc("/healthz", writeOKHealth)
	mux.HandleFunc("/api/health", writeOKHealth)
	mux.HandleFunc("/api/healthz", writeOKHealth)

	mux.HandleFunc("/api/greeting", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		writeJSON(w, http.StatusOK, map[string]any{
			"greeting":    greeting,
			"build_label": buildLabel,
			"git_sha":     buildSHA,
			"git_ref":     buildRef,
			"version":     appVersion,
			"s3_ready":    s3OK,
		})
	})

	mux.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		if !s3OK || s3c == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": "S3 chưa sẵn sàng — bật MinIO addon (inject S3_ENDPOINT / S3_ACCESS_KEY / S3_SECRET_KEY / S3_BUCKET)",
			})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		switch r.Method {
		case http.MethodGet:
			out, err := s3c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:  aws.String(s3cfg.Bucket),
				MaxKeys: aws.Int32(200),
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
			items := make([]map[string]any, 0, len(out.Contents))
			for _, obj := range out.Contents {
				items = append(items, map[string]any{
					"key":           aws.ToString(obj.Key),
					"size":          aws.ToInt64(obj.Size),
					"last_modified": obj.LastModified,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"bucket": s3cfg.Bucket,
				"items":  items,
			})

		case http.MethodPost:
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "multipart không hợp lệ: " + err.Error()})
				return
			}
			file, hdr, err := r.FormFile("file")
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu field file"})
				return
			}
			defer file.Close()
			key := strings.TrimSpace(r.FormValue("key"))
			if key == "" {
				key = path.Base(hdr.Filename)
			}
			key, err = sanitizeObjectKey(key)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			ct := hdr.Header.Get("Content-Type")
			if ct == "" {
				ct = "application/octet-stream"
			}
			_, err = s3c.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(s3cfg.Bucket),
				Key:         aws.String(key),
				Body:        file,
				ContentType: aws.String(ct),
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":     true,
				"bucket": s3cfg.Bucket,
				"key":    key,
				"size":   hdr.Size,
			})

		case http.MethodDelete:
			key, err := sanitizeObjectKey(r.URL.Query().Get("key"))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			_, err = s3c.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s3cfg.Bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "key": key})

		default:
			w.Header().Set("Allow", "GET, POST, DELETE")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method không hỗ trợ"})
		}
	})

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/" || r.URL.Path == "/api" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = fmt.Fprintf(w, "test-k8s api v%s\n", appVersion)
			return
		}
		http.NotFound(w, r)
	})

	log.Printf("api listening on :%s (%s)", port, appVersion)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
