package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/rabbitmq"
)

// Handler handles HTTP requests
type Handler struct {
	manager *rabbitmq.Manager
}

// NewHandler creates a new API handler
func NewHandler(manager *rabbitmq.Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/connections", h.ListConnections).Methods("GET")
	api.HandleFunc("/vhosts", h.ListVHosts).Methods("GET")
	api.HandleFunc("/streams", h.ListStreams).Methods("GET")
	api.HandleFunc("/streams/{connection_id}/{vhost}/{stream_name}/stats", h.GetStreamStats).Methods("GET")
	api.HandleFunc("/streams/{connection_id}/{vhost}/{stream_name}/messages", h.GetMessages).Methods("GET")

	// Health check
	r.HandleFunc("/health", h.Health).Methods("GET")
}

// ListConnections returns all configured connections
func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	connections := h.manager.ListConnections()
	respondJSON(w, http.StatusOK, connections)
}

// ListVHosts returns all vhosts across all connections
func (h *Handler) ListVHosts(w http.ResponseWriter, r *http.Request) {
	vhosts, err := h.manager.ListVHosts(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list vhosts", err)
		return
	}

	respondJSON(w, http.StatusOK, vhosts)
}

// ListStreams returns all streams across all connections
func (h *Handler) ListStreams(w http.ResponseWriter, r *http.Request) {
	streams, err := h.manager.ListStreams(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list streams", err)
		return
	}

	respondJSON(w, http.StatusOK, streams)
}

// GetStreamStats returns statistics for a specific stream
func (h *Handler) GetStreamStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionID := vars["connection_id"]
	vhost := vars["vhost"]
	streamName := vars["stream_name"]

	conn, err := h.manager.GetConnection(connectionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Connection not found", err)
		return
	}

	stats, err := conn.GetStreamStatsForVHost(r.Context(), vhost, streamName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stream stats", err)
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetMessages returns messages from a stream
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionID := vars["connection_id"]
	vhost := vars["vhost"]
	streamName := vars["stream_name"]

	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := uint64(0)
	if offsetStr != "" {
		parsed, err := strconv.ParseUint(offsetStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid offset parameter", err)
			return
		}
		offset = parsed
	}

	limit := 10
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid limit parameter", err)
			return
		}
		limit = parsed
	}

	conn, err := h.manager.GetConnection(connectionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Connection not found", err)
		return
	}

	messages, err := conn.ReadMessagesFromVHost(r.Context(), vhost, streamName, offset, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read messages", err)
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

// Health returns the health status of the service
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]string{
		"error":   message,
		"details": err.Error(),
	}
	respondJSON(w, status, response)
}

