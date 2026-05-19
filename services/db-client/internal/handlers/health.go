package handlers

import (
		"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler{
	return &HealthHandler{pool: pool}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	log.Println("HealthHandler.Health")
	w.Header().Set("Content-Type", "application/json")
	
	if h.pool.Ping(r.Context()) != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"bad"}`))
		return
	}
	
	w.Write([]byte(`{"status":"ok"}`))
}
