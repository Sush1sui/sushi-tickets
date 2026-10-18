// internal/api/middleware.go
package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/jackc/pgx/v5"
)

// SecurityHeaders adds hardening HTTP headers to every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

const maxBodySize = 10 * 1024 * 1024 // 10 MB

// LimitRequestBody restricts request body size to 10 MB.
func LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		}
		next.ServeHTTP(w, r)
	})
}

// EnableCORS allows your React frontend to communicate with this Go API.
// It also handles Chrome's Private Network Access (PNA) preflight:
// https://developer.chrome.com/blog/private-network-access-preflight/
func EnableCORS(next http.Handler, allowedOrigin string, allowCredentials bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}

		if allowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-CSRF-Token")

		// Chrome Private Network Access: echo back the PNA header on preflight
		// so the browser doesn't show the "access local network" permission prompt.
		if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

const (
	idempotencyHeader       = "Idempotency-Key"
	idempotencyTTL          = 10 * time.Minute
	maxIdempotencyKeyLength = 200
	maxIdempotencyBodySize  = 1024 * 1024
)

var errIdempotencyConflict = errors.New("idempotency key conflict")

type idempotencyContext struct {
	key         string
	userID      string
	method      string
	path        string
	requestHash string
	createdAt   int64
	expiresAt   int64
}

type responseCapture struct {
	writer     http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	truncated  bool
}

func (r *responseCapture) Header() http.Header {
	return r.writer.Header()
}

func (r *responseCapture) WriteHeader(statusCode int) {
	if r.statusCode == 0 {
		r.statusCode = statusCode
	}
	r.writer.WriteHeader(statusCode)
}

func (r *responseCapture) Write(data []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	if !r.truncated {
		remaining := maxIdempotencyBodySize - r.body.Len()
		if remaining > 0 {
			if len(data) > remaining {
				_, _ = r.body.Write(data[:remaining])
				r.truncated = true
			} else {
				_, _ = r.body.Write(data)
			}
		} else {
			r.truncated = true
		}
	}
	return r.writer.Write(data)
}

func isIdempotentMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func isCSRFProtectedMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func validateCSRF(r *http.Request) bool {
	headerToken := strings.TrimSpace(r.Header.Get(csrfHeaderName))
	if headerToken == "" {
		return false
	}
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) == 1
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{buckets: make(map[string]*rateBucket)}
}

func (r *rateLimiter) Allow(key string, limit int, window time.Duration) bool {
	if key == "" {
		return true
	}

	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket := r.buckets[key]
	if bucket == nil || now.After(bucket.resetAt) {
		r.buckets[key] = &rateBucket{count: 1, resetAt: now.Add(window)}
		return true
	}

	if bucket.count >= limit {
		return false
	}
	bucket.count++
	return true
}

func (r *rateLimiter) startCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			r.mu.Lock()
			now := time.Now()
			for key, bucket := range r.buckets {
				if now.After(bucket.resetAt) {
					delete(r.buckets, key)
				}
			}
			r.mu.Unlock()
		}
	}()
}

func (s *Server) clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" && s.isTrustedProxy(r.RemoteAddr) {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return stripPort(r.RemoteAddr)
}

func (s *Server) isTrustedProxy(addr string) bool {
	host := stripPort(addr)
	for _, proxy := range s.Config.TrustedProxies {
		if host == proxy {
			return true
		}
	}
	return false
}

func stripPort(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func (s *Server) startIdempotency(w http.ResponseWriter, r *http.Request, claims *AuthClaims) (*idempotencyContext, *responseCapture, error) {
	if claims == nil || claims.UserID == "" {
		return nil, nil, errors.New("missing user")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	bodyHash := sha256.Sum256(bodyBytes)
	requestHash := hex.EncodeToString(bodyHash[:])

	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path = path + "?" + r.URL.RawQuery
	}

	key := strings.TrimSpace(r.Header.Get(idempotencyHeader))
	if key == "" {
		autoKey := sha256.Sum256([]byte(claims.UserID + "|" + r.Method + "|" + path + "|" + requestHash))
		key = hex.EncodeToString(autoKey[:])
	}
	if len(key) > maxIdempotencyKeyLength {
		return nil, nil, errors.New("idempotency key too long")
	}

	now := time.Now().Unix()
	record, err := s.DB.GetIdempotencyRecord(r.Context(), key, claims.UserID)
	if err == nil {
		if record.ExpiresAt < now {
			_ = s.DB.DeleteIdempotencyRecord(r.Context(), key, claims.UserID)
		} else if record.RequestHash != requestHash || record.Method != r.Method || record.Path != path {
			return nil, nil, errIdempotencyConflict
		} else {
			contentType := record.ResponseContentType
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			w.WriteHeader(int(record.ResponseCode))
			_, _ = w.Write(record.ResponseBody)
			return nil, nil, nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, err
	}

	ctx := &idempotencyContext{
		key:         key,
		userID:      claims.UserID,
		method:      r.Method,
		path:        path,
		requestHash: requestHash,
		createdAt:   now,
		expiresAt:   now + int64(idempotencyTTL.Seconds()),
	}

	return ctx, &responseCapture{writer: w}, nil
}

func (s *Server) finishIdempotency(requestCtx context.Context, meta *idempotencyContext, recorder *responseCapture) error {
	if meta == nil || recorder == nil {
		return nil
	}
	if recorder.statusCode == 0 {
		recorder.statusCode = http.StatusOK
	}
	if recorder.truncated || recorder.statusCode >= http.StatusInternalServerError {
		return nil
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	return s.DB.CreateIdempotencyRecord(requestCtx, db.CreateIdempotencyRecordParams{
		Key:                 meta.key,
		UserID:              meta.userID,
		Method:              meta.method,
		Path:                meta.path,
		RequestHash:         meta.requestHash,
		ResponseCode:        int32(recorder.statusCode),
		ResponseBody:        recorder.body.Bytes(),
		ResponseContentType: contentType,
		CreatedAt:           meta.createdAt,
		ExpiresAt:           meta.expiresAt,
	})
}
