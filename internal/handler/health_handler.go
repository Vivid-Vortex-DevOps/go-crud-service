package handler

import (
	"context"
	"net/http"
	"time"
)

// DBPinger is implemented by pgxpool.Pool.
type DBPinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db DBPinger
}

func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live always returns 200 — the process is running.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns 200 when the database is reachable, 503 otherwise.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	dbStatus := "ok"
	status := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "unavailable"
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]any{
		"status": map[string]string{"ok": "ready", "unavailable": "unavailable"}[dbStatus],
		"checks": map[string]string{"database": dbStatus},
	})
}
