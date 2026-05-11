package handlers

import (
	"db-client/internal/db"
	"net/http"
)

type HealthHandler struct {
	db *db.DBClient
}

func NewHealthHandler(db *db.DBClient) *HealthHandler{
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.db.Ping() != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"bad"}`))
		return
	}
	
	w.Write([]byte(`{"status":"ok"}`))
}