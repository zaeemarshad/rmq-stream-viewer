package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Connections []ConnectionConfig `yaml:"connections"`
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Port int `yaml:"port"`
}

// ConnectionConfig represents a RabbitMQ connection
type ConnectionConfig struct {
	ID         string `yaml:"id" json:"id"`
	Name       string `yaml:"name" json:"name"`
	Host       string `yaml:"host" json:"host"`
	Port       int    `yaml:"port" json:"port"`
	VHost      string `yaml:"vhost" json:"vhost"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	HTTPPort   int    `yaml:"http_port" json:"http_port"`     // For management API
	StreamPort int    `yaml:"stream_port" json:"stream_port"` // For stream protocol (default: 5552)
}

// AMQPURL returns the AMQP connection URL
func (c *ConnectionConfig) AMQPURL() string {
	vhost := c.VHost
	if vhost == "" {
		vhost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, vhost)
}

// ManagementURL returns the RabbitMQ management API URL
func (c *ConnectionConfig) ManagementURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.HTTPPort)
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}

	if len(c.Connections) == 0 {
		return fmt.Errorf("at least one connection must be configured")
	}

	seen := make(map[string]bool)
	for i, conn := range c.Connections {
		if conn.ID == "" {
			return fmt.Errorf("connection %d: ID is required", i)
		}
		if seen[conn.ID] {
			return fmt.Errorf("connection %d: duplicate ID '%s'", i, conn.ID)
		}
		seen[conn.ID] = true

		if conn.Host == "" {
			return fmt.Errorf("connection '%s': host is required", conn.ID)
		}
		if conn.Port <= 0 {
			return fmt.Errorf("connection '%s': port must be positive", conn.ID)
		}
		if conn.HTTPPort <= 0 {
			return fmt.Errorf("connection '%s': http_port must be positive", conn.ID)
		}
		if conn.Username == "" {
			return fmt.Errorf("connection '%s': username is required", conn.ID)
		}

		// Set default stream port if not specified
		if c.Connections[i].StreamPort <= 0 {
			c.Connections[i].StreamPort = 5552
		}
	}

	return nil
}

