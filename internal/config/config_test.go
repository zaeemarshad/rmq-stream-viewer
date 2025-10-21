package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	validConfig := `
server:
  port: 8080
connections:
  - id: conn1
    name: Development
    host: localhost
    port: 5672
    vhost: /
    username: guest
    password: guest
    http_port: 15672
`

	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}

	if len(cfg.Connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(cfg.Connections))
	}

	conn := cfg.Connections[0]
	if conn.ID != "conn1" {
		t.Errorf("Expected ID 'conn1', got '%s'", conn.ID)
	}
	if conn.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", conn.Host)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Connections: []ConnectionConfig{
					{
						ID:       "conn1",
						Name:     "Test",
						Host:     "localhost",
						Port:     5672,
						HTTPPort: 15672,
						Username: "guest",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid server port",
			config: Config{
				Server: ServerConfig{Port: 0},
				Connections: []ConnectionConfig{
					{ID: "conn1", Host: "localhost", Port: 5672, HTTPPort: 15672, Username: "guest"},
				},
			},
			wantErr: true,
		},
		{
			name: "no connections",
			config: Config{
				Server:      ServerConfig{Port: 8080},
				Connections: []ConnectionConfig{},
			},
			wantErr: true,
		},
		{
			name: "duplicate connection IDs",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Connections: []ConnectionConfig{
					{ID: "conn1", Host: "localhost", Port: 5672, HTTPPort: 15672, Username: "guest"},
					{ID: "conn1", Host: "localhost", Port: 5672, HTTPPort: 15672, Username: "guest"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAMQPURL(t *testing.T) {
	conn := ConnectionConfig{
		Host:     "localhost",
		Port:     5672,
		VHost:    "/test",
		Username: "user",
		Password: "pass",
	}

	expected := "amqp://user:pass@localhost:5672//test"
	if got := conn.AMQPURL(); got != expected {
		t.Errorf("AMQPURL() = %v, want %v", got, expected)
	}

	// Test default vhost
	conn.VHost = ""
	expected = "amqp://user:pass@localhost:5672//"
	if got := conn.AMQPURL(); got != expected {
		t.Errorf("AMQPURL() with default vhost = %v, want %v", got, expected)
	}
}

