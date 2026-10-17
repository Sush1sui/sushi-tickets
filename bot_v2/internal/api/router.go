package api

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Sush1sui/FNS_BOT/internal/api/auth"
	"github.com/Sush1sui/FNS_BOT/internal/api/panels"
	serverconfig "github.com/Sush1sui/FNS_BOT/internal/api/server-config"
	"github.com/Sush1sui/FNS_BOT/internal/api/transcripts"
	"github.com/Sush1sui/FNS_BOT/internal/config"
	"github.com/Sush1sui/FNS_BOT/internal/db"
	"github.com/Sush1sui/FNS_BOT/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(queries *db.Queries, storageClient *storage.Client, cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	s := &Server{
		DB:      queries,
		Storage: storageClient,
		Config:  cfg,
		Limiter: newRateLimiter(),
	}
	s.Limiter.startCleanup(5 * time.Minute)
	s.startSessionCleanup(4 * time.Hour)

	configHandler := &serverconfig.Handler{DB: queries}
	authHandler := &auth.Handler{Server: s}
	panelsHandler := &panels.Handler{DB: queries, Pool: pool}
	transcriptsHandler := &transcripts.Handler{DB: queries, Storage: s.Storage}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ip := s.clientIP(r)
		if !s.Limiter.Allow("health:"+ip, 30, time.Minute) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
			return
		}
		s.handleHealth(w, r)
	})

	// Server config routes with auth wrapper
	mux.HandleFunc("GET /api/config/{server_id}", s.wrapAuthConfig(configHandler.HandleGetServerConfig))
	mux.HandleFunc("PUT /api/config/{server_id}", s.wrapAuthConfig(configHandler.HandleUpdateServerConfig))
	mux.HandleFunc("GET /api/servers/{server_id}/meta", s.wrapAuthConfig(configHandler.HandleGetMeta))
	mux.HandleFunc("GET /api/servers/{server_id}/meta/roles", s.wrapAuthConfig(configHandler.HandleGetRoles))
	mux.HandleFunc("GET /api/servers/{server_id}/meta/channels", s.wrapAuthConfig(configHandler.HandleGetChannels))
	mux.HandleFunc("GET /api/servers/{server_id}/meta/categories", s.wrapAuthConfig(configHandler.HandleGetCategories))
	mux.HandleFunc("GET /api/servers/{server_id}/meta/emojis", s.wrapAuthConfig(configHandler.HandleGetEmojis))

	// Staff routes
	mux.HandleFunc("GET /api/servers/{server_id}/staff", s.wrapAuthConfig(configHandler.HandleGetStaff))
	mux.HandleFunc("PUT /api/servers/{server_id}/staff", s.wrapAuthConfig(configHandler.HandleUpdateStaff))

	// Auth routes
	mux.HandleFunc("GET /api/auth/login", s.wrapIPRateLimit("auth:login:", 20, time.Minute, authHandler.HandleAuthLogin))
	mux.HandleFunc("GET /api/auth/callback", s.wrapIPRateLimit("auth:callback:", 40, time.Minute, authHandler.HandleAuthCallback))
	mux.HandleFunc("GET /api/auth/me", s.wrapIPRateLimit("auth:me:", 60, time.Minute, authHandler.HandleAuthMe))
	mux.HandleFunc("GET /api/auth/servers", s.wrapIPRateLimit("auth:servers:", 60, time.Minute, authHandler.HandleAuthServers))
	mux.HandleFunc("POST /api/auth/logout", s.wrapIPRateLimit("auth:logout:", 30, time.Minute, authHandler.HandleAuthLogout))

	// Panel routes
	mux.HandleFunc("GET /api/servers/{server_id}/panels", s.wrapAuthConfig(panelsHandler.HandleListPanels))
	mux.HandleFunc("GET /api/servers/{server_id}/panels/{panel_id}", s.wrapAuthConfig(panelsHandler.HandleGetPanel))
	mux.HandleFunc("POST /api/servers/{server_id}/panels", s.wrapAuthConfig(panelsHandler.HandleCreatePanel))
	mux.HandleFunc("PUT /api/servers/{server_id}/panels/{panel_id}", s.wrapAuthConfig(panelsHandler.HandleUpdatePanel))
	mux.HandleFunc("DELETE /api/servers/{server_id}/panels/{panel_id}", s.wrapAuthConfig(panelsHandler.HandleDeletePanel))
	mux.HandleFunc("POST /api/servers/{server_id}/panels/{panel_id}/send", s.wrapAuthConfig(panelsHandler.HandleSendPanel))

	// Panel schedule window routes
	mux.HandleFunc("GET /api/servers/{server_id}/panels/{panel_id}/schedule", s.wrapAuthConfig(panelsHandler.HandleListScheduleWindows))
	mux.HandleFunc("POST /api/servers/{server_id}/panels/{panel_id}/schedule", s.wrapAuthConfig(panelsHandler.HandleCreateScheduleWindow))
	mux.HandleFunc("PUT /api/servers/{server_id}/panels/{panel_id}/schedule/{window_id}", s.wrapAuthConfig(panelsHandler.HandleUpdateScheduleWindow))
	mux.HandleFunc("DELETE /api/servers/{server_id}/panels/{panel_id}/schedule/{window_id}", s.wrapAuthConfig(panelsHandler.HandleDeleteScheduleWindow))
	mux.HandleFunc("POST /api/servers/{server_id}/panels/{panel_id}/schedule/{window_id}/close", s.wrapAuthConfig(panelsHandler.HandleCloseScheduleWindow))
	mux.HandleFunc("POST /api/servers/{server_id}/panels/{panel_id}/schedule/{window_id}/open", s.wrapAuthConfig(panelsHandler.HandleReopenScheduleWindow))

	// Multi-panel routes
	mux.HandleFunc("GET /api/servers/{server_id}/multi-panels", s.wrapAuthConfig(panelsHandler.HandleListMultiPanels))
	mux.HandleFunc("GET /api/servers/{server_id}/multi-panels/{multi_panel_id}", s.wrapAuthConfig(panelsHandler.HandleGetMultiPanel))
	mux.HandleFunc("POST /api/servers/{server_id}/multi-panels", s.wrapAuthConfig(panelsHandler.HandleCreateMultiPanel))
	mux.HandleFunc("PUT /api/servers/{server_id}/multi-panels/{multi_panel_id}", s.wrapAuthConfig(panelsHandler.HandleUpdateMultiPanel))
	mux.HandleFunc("DELETE /api/servers/{server_id}/multi-panels/{multi_panel_id}", s.wrapAuthConfig(panelsHandler.HandleDeleteMultiPanel))
	mux.HandleFunc("POST /api/servers/{server_id}/multi-panels/{multi_panel_id}/send", s.wrapAuthConfig(panelsHandler.HandleSendMultiPanel))

	// Transcript routes
	mux.HandleFunc("GET /api/servers/{server_id}/transcripts", s.wrapAuthConfig(transcriptsHandler.HandleListTranscripts))
	mux.HandleFunc("GET /api/servers/{server_id}/transcripts/{transcript_id}", s.wrapAuthConfig(transcriptsHandler.HandleGetTranscript))
	mux.HandleFunc("GET /api/servers/{server_id}/transcripts/{transcript_id}/content", s.wrapAuthConfig(transcriptsHandler.HandleGetTranscriptContent))

	return SecurityHeaders(LimitRequestBody(EnableCORS(mux, cfg.ClientOrigin, true)))
}

// wrapAuthConfig wraps a handler to require auth + server authorization
func (s *Server) wrapAuthConfig(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serverID := r.PathValue("server_id")

		claims, err := s.AuthFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		authorized, err := s.IsAuthorizedForServer(r.Context(), serverID, claims)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "auth check failed"})
			return
		}
		if !authorized {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
			return
		}

		if s.Limiter != nil {
			key := "user:" + claims.UserID
			if !s.Limiter.Allow(key, 60, time.Minute) {
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
				return
			}
		}

		if isCSRFProtectedMethod(r.Method) && !validateCSRF(r) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "csrf"})
			return
		}

		if isIdempotentMethod(r.Method) {
			ctx, recorder, err := s.startIdempotency(w, r, claims)
			if err != nil {
				if errors.Is(err, errIdempotencyConflict) {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "idempotency key conflict"})
					return
				}
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "idempotency check failed"})
				return
			}
			if recorder == nil {
				return
			}
			next(recorder, r)
			if err := s.finishIdempotency(r.Context(), ctx, recorder); err != nil {
				log.Printf("idempotency store failed: %v", err)
			}
			return
		}

		next(w, r)
	}
}

func (s *Server) wrapIPRateLimit(prefix string, limit int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Limiter != nil {
			ip := s.clientIP(r)
			if !s.Limiter.Allow(prefix+ip, limit, window) {
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
				return
			}
		}
		next(w, r)
	}
}
