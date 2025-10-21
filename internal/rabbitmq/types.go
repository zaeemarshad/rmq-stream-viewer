package rabbitmq

import "time"

// VHost represents a virtual host
type VHost struct {
	Name         string   `json:"name"`
	ConnectionID string   `json:"connection_id"`
	Streams      []Stream `json:"streams,omitempty"`
}

// Stream represents a RabbitMQ stream
type Stream struct {
	Name         string `json:"name"`
	ConnectionID string `json:"connection_id"`
	VHost        string `json:"vhost"`
}

// StreamStats represents statistics for a stream
type StreamStats struct {
	Name        string `json:"name"`
	MessageCount int64  `json:"message_count"`
	Size        int64  `json:"size"`
	FirstOffset uint64 `json:"first_offset"`
	LastOffset  uint64 `json:"last_offset"`
}

// Message represents a message from a stream
type Message struct {
	Offset     uint64                 `json:"offset"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       []byte                 `json:"data"`
	Properties map[string]interface{} `json:"properties"`
}

// MessageBatch represents a batch of messages with metadata
type MessageBatch struct {
	Messages   []Message `json:"messages"`
	StartOffset uint64    `json:"start_offset"`
	EndOffset   uint64    `json:"end_offset"`
	HasMore     bool      `json:"has_more"`
}

