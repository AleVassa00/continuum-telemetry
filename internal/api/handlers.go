package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"continuum-telemetry/internal/storage"
)

type Server struct {
	store storage.AggregateReader
}

func NewServer(store storage.AggregateReader) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /stats", s.handleStats)
	mux.HandleFunc("GET /aggregates/latest", s.handleLatestAggregates)
	mux.HandleFunc("GET /aggregates", s.handleAggregatesBySensor)

	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	count, err := s.store.CountAggregates(r.Context())
	if err != nil {
		log.Printf("stats_error=%v", err)
		http.Error(w, "failed to read stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stored_aggregates": count,
	})
}

func (s *Server) handleLatestAggregates(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 20)

	aggregates, err := s.store.ListLatestAggregates(r.Context(), limit)
	if err != nil {
		log.Printf("latest_aggregates_error=%v", err)
		http.Error(w, "failed to read latest aggregates", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":      len(aggregates),
		"aggregates": aggregates,
	})
}

func (s *Server) handleAggregatesBySensor(w http.ResponseWriter, r *http.Request) {
	sensorID := r.URL.Query().Get("sensor_id")
	if sensorID == "" {
		http.Error(w, "missing required query parameter: sensor_id", http.StatusBadRequest)
		return
	}

	limit := parseLimit(r, 20)

	aggregates, err := s.store.ListAggregatesBySensor(r.Context(), sensorID, limit)
	if err != nil {
		log.Printf("sensor_aggregates_error sensor_id=%s error=%v", sensorID, err)
		http.Error(w, "failed to read sensor aggregates", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sensor_id":  sensorID,
		"count":      len(aggregates),
		"aggregates": aggregates,
	})
}

func parseLimit(r *http.Request, fallback int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return fallback
	}

	if limit > 100 {
		return 100
	}

	return limit
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write_json_error=%v", err)
	}
}
