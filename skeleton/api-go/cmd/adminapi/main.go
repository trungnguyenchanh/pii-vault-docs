package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/pii-admin-api/internal/audit"
	"github.com/your-org/pii-admin-api/internal/auth"
	"github.com/your-org/pii-admin-api/internal/config"
	"github.com/your-org/pii-admin-api/internal/handler"
	"github.com/your-org/pii-admin-api/internal/repo"
)

func main() {
	cfg := config.Load()

	// Kết nối DB (read-only credentials cho audit). Cho phép chạy không DB ở dev.
	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Printf("warn: cannot connect DB: %v (chạy chế độ skeleton)", err)
		} else {
			pool = p
			defer pool.Close()
		}
	}

	rp := repo.New(pool)
	rec := audit.NewRecorder()
	h := handler.New(rp, rec)
	authn := auth.NewAuthenticator(cfg.JWKSURL, cfg.JWTIssuer, cfg.JWTAudience)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Debug-User", "X-Debug-Roles"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health (không cần auth).
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Toàn bộ API yêu cầu xác thực + vai trò admin/dpo/security.
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(authn.Middleware)
		r.Use(auth.RequireRole(auth.RoleAdmin, auth.RoleDPO, auth.RoleSecurity))

		// F1 — Logs
		r.Get("/logs", h.ListLogs)
		r.Get("/logs/{seq}", h.GetLog)
		r.Post("/logs/verify-chain", h.VerifyChain)

		// F2 — Stats
		r.Get("/stats/access", h.StatsAccess)
		r.Get("/stats/top-actors", h.StatsTopActors)
		r.Get("/stats/summary", h.StatsSummary)

		// F3 — Alerts & investigation
		r.Get("/alerts", h.ListAlerts)
		r.Patch("/alerts/{id}", h.UpdateAlert)
		r.Get("/subjects/{ref}/timeline", h.SubjectTimeline)
		r.Post("/actors/{id}/revoke", h.RevokeActor)

		// F4 — Approvals (four-eyes)
		r.Get("/approvals", h.ListApprovals)
		r.Post("/approvals/{id}/approve", h.ApproveRequest)
		r.Post("/approvals/{id}/reject", h.RejectRequest)
	})

	log.Printf("Admin API listening on %s (dev mode: %v)", cfg.Addr, cfg.JWKSURL == "")
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
