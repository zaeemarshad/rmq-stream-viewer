package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/config"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/rabbitmq"
)

func TestHealth(t *testing.T) {
	manager := rabbitmq.NewManager([]config.ConnectionConfig{})
	handler := NewHandler(manager)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("unexpected response: got %v want ok", response["status"])
	}
}

func TestListConnections(t *testing.T) {
	configs := []config.ConnectionConfig{
		{
			ID:       "test1",
			Name:     "Test Connection 1",
			Host:     "localhost",
			Port:     5672,
			HTTPPort: 15672,
			Username: "guest",
		},
	}

	manager := rabbitmq.NewManager(configs)
	handler := NewHandler(manager)

	req, err := http.NewRequest("GET", "/api/connections", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response []config.ConnectionConfig
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 connection, got %d", len(response))
	}

	if response[0].ID != "test1" {
		t.Errorf("unexpected connection ID: got %v want test1", response[0].ID)
	}
}

func TestGetMessages_InvalidOffset(t *testing.T) {
	manager := rabbitmq.NewManager([]config.ConnectionConfig{})
	handler := NewHandler(manager)

	req, err := http.NewRequest("GET", "/api/streams/conn1/stream1/messages?offset=invalid", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetMessages_ConnectionNotFound(t *testing.T) {
	manager := rabbitmq.NewManager([]config.ConnectionConfig{})
	handler := NewHandler(manager)

	req, err := http.NewRequest("GET", "/api/streams/nonexistent/stream1/messages", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

