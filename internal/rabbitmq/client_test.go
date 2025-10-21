package rabbitmq

import (
	"testing"

	"github.com/zaeem.arshad/rmq-stream-viewer/internal/config"
)

func TestNewManager(t *testing.T) {
	configs := []config.ConnectionConfig{
		{
			ID:       "test1",
			Name:     "Test 1",
			Host:     "localhost",
			Port:     5672,
			VHost:    "/",
			Username: "guest",
			Password: "guest",
			HTTPPort: 15672,
		},
	}

	manager := NewManager(configs)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if len(manager.configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(manager.configs))
	}

	if len(manager.connections) != 0 {
		t.Errorf("Expected 0 connections before Connect(), got %d", len(manager.connections))
	}
}

func TestListConnections(t *testing.T) {
	configs := []config.ConnectionConfig{
		{
			ID:       "test1",
			Name:     "Test 1",
			Host:     "localhost",
			Port:     5672,
			VHost:    "/",
			Username: "guest",
			Password: "guest",
			HTTPPort: 15672,
		},
		{
			ID:       "test2",
			Name:     "Test 2",
			Host:     "localhost",
			Port:     5672,
			VHost:    "/test",
			Username: "guest",
			Password: "guest",
			HTTPPort: 15672,
		},
	}

	manager := NewManager(configs)
	connections := manager.ListConnections()

	if len(connections) != 2 {
		t.Errorf("Expected 2 connections, got %d", len(connections))
	}

	if connections[0].ID != "test1" {
		t.Errorf("Expected first connection ID 'test1', got '%s'", connections[0].ID)
	}

	if connections[1].ID != "test2" {
		t.Errorf("Expected second connection ID 'test2', got '%s'", connections[1].ID)
	}
}

func TestGetConnection_NotFound(t *testing.T) {
	manager := NewManager([]config.ConnectionConfig{})

	_, err := manager.GetConnection("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent connection, got nil")
	}

	expectedMsg := "connection not found"
	if err.Error() != "connection not found: nonexistent" {
		t.Errorf("Expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestMessageBatch(t *testing.T) {
	batch := &MessageBatch{
		Messages: []Message{
			{
				Offset:    100,
				Data:      []byte("test"),
				Properties: map[string]interface{}{
					"test": "value",
				},
			},
		},
		StartOffset: 100,
		EndOffset:   100,
		HasMore:     false,
	}

	if len(batch.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(batch.Messages))
	}

	if batch.StartOffset != 100 {
		t.Errorf("Expected StartOffset 100, got %d", batch.StartOffset)
	}

	if batch.EndOffset != 100 {
		t.Errorf("Expected EndOffset 100, got %d", batch.EndOffset)
	}

	if batch.HasMore {
		t.Error("Expected HasMore false, got true")
	}
}

func TestStream(t *testing.T) {
	stream := Stream{
		Name:         "test-stream",
		ConnectionID: "conn1",
		VHost:        "/",
	}

	if stream.Name != "test-stream" {
		t.Errorf("Expected Name 'test-stream', got '%s'", stream.Name)
	}

	if stream.ConnectionID != "conn1" {
		t.Errorf("Expected ConnectionID 'conn1', got '%s'", stream.ConnectionID)
	}

	if stream.VHost != "/" {
		t.Errorf("Expected VHost '/', got '%s'", stream.VHost)
	}
}

func TestStreamStats(t *testing.T) {
	stats := StreamStats{
		Name:         "test-stream",
		MessageCount: 1000,
		Size:         1024000,
		FirstOffset:  0,
		LastOffset:   999,
	}

	if stats.Name != "test-stream" {
		t.Errorf("Expected Name 'test-stream', got '%s'", stats.Name)
	}

	if stats.MessageCount != 1000 {
		t.Errorf("Expected MessageCount 1000, got %d", stats.MessageCount)
	}

	if stats.Size != 1024000 {
		t.Errorf("Expected Size 1024000, got %d", stats.Size)
	}

	if stats.FirstOffset != 0 {
		t.Errorf("Expected FirstOffset 0, got %d", stats.FirstOffset)
	}

	if stats.LastOffset != 999 {
		t.Errorf("Expected LastOffset 999, got %d", stats.LastOffset)
	}
}

func TestMessage(t *testing.T) {
	props := map[string]interface{}{
		"message_id":   "test-123",
		"content_type": "application/json",
		"counter":      42,
	}

	msg := Message{
		Offset:     100,
		Data:       []byte(`{"test": "value"}`),
		Properties: props,
	}

	if msg.Offset != 100 {
		t.Errorf("Expected Offset 100, got %d", msg.Offset)
	}

	expectedData := `{"test": "value"}`
	if string(msg.Data) != expectedData {
		t.Errorf("Expected Data '%s', got '%s'", expectedData, string(msg.Data))
	}

	if len(msg.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(msg.Properties))
	}

	if msg.Properties["message_id"] != "test-123" {
		t.Errorf("Expected message_id 'test-123', got '%v'", msg.Properties["message_id"])
	}

	if msg.Properties["counter"] != 42 {
		t.Errorf("Expected counter 42, got '%v'", msg.Properties["counter"])
	}
}

