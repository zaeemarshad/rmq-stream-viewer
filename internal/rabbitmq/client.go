package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"github.com/zaeem.arshad/rmq-stream-viewer/internal/config"
)

// Manager manages multiple RabbitMQ connections
type Manager struct {
	connections map[string]*Connection
	configs     []config.ConnectionConfig
	mu          sync.RWMutex
}

// Connection represents a RabbitMQ connection
type Connection struct {
	ID          string
	Name        string
	Config      config.ConnectionConfig
	Environment *stream.Environment
	httpClient  *http.Client
}

// NewManager creates a new RabbitMQ connection manager
func NewManager(configs []config.ConnectionConfig) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		configs:     configs,
	}
}

// Connect establishes connections to all configured RabbitMQ instances
func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cfg := range m.configs {
		vhost := cfg.VHost
		if vhost == "" {
			vhost = "/"
		}

		env, err := stream.NewEnvironment(
			stream.NewEnvironmentOptions().
				SetHost(cfg.Host).
				SetPort(cfg.StreamPort).
				SetUser(cfg.Username).
				SetPassword(cfg.Password).
				SetVHost(vhost),
		)
		if err != nil {
			return fmt.Errorf("failed to create environment for %s: %w", cfg.ID, err)
		}

		conn := &Connection{
			ID:          cfg.ID,
			Name:        cfg.Name,
			Config:      cfg,
			Environment: env,
			httpClient:  &http.Client{Timeout: 10 * time.Second},
		}

		m.connections[cfg.ID] = conn
	}

	return nil
}

// Close closes all connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, conn := range m.connections {
		if err := conn.Environment.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// GetConnection returns a connection by ID
func (m *Manager) GetConnection(id string) (*Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, ok := m.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	return conn, nil
}

// ListConnections returns all configured connections
func (m *Manager) ListConnections() []config.ConnectionConfig {
	return m.configs
}

// ListVHosts returns all vhosts across all connections with their streams
func (m *Manager) ListVHosts(ctx context.Context) ([]VHost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allVHosts []VHost

	for _, conn := range m.connections {
		vhosts, err := conn.ListVHosts(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list vhosts for %s: %w", conn.ID, err)
		}
		allVHosts = append(allVHosts, vhosts...)
	}

	return allVHosts, nil
}

// ListStreams returns all streams across all connections
func (m *Manager) ListStreams(ctx context.Context) ([]Stream, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allStreams []Stream

	for _, conn := range m.connections {
		streams, err := conn.ListStreams(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list streams for %s: %w", conn.ID, err)
		}
		allStreams = append(allStreams, streams...)
	}

	return allStreams, nil
}

// ListVHosts returns all vhosts for this connection with their streams
func (c *Connection) ListVHosts(ctx context.Context) ([]VHost, error) {
	//Query Management API for vhosts
	apiURL := fmt.Sprintf("%s/api/vhosts", c.Config.ManagementURL())
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query management API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("management API returned %d: %s", resp.StatusCode, string(body))
	}

	var apiVHosts []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiVHosts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var vhosts []VHost
	for _, v := range apiVHosts {
		streams, err := c.ListStreamsInVHost(ctx, v.Name)
		if err != nil {
			// Skip vhosts that fail to list streams
			continue
		}
		
		vhosts = append(vhosts, VHost{
			Name:         v.Name,
			ConnectionID: c.ID,
			Streams:      streams,
		})
	}

	return vhosts, nil
}

// ListStreamsInVHost returns all streams in a specific vhost
func (c *Connection) ListStreamsInVHost(ctx context.Context, vhost string) ([]Stream, error) {
	apiURL := fmt.Sprintf("%s/api/queues/%s", c.Config.ManagementURL(), url.PathEscape(vhost))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query management API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("management API returned %d: %s", resp.StatusCode, string(body))
	}

	var apiQueues []struct {
		Name  string `json:"name"`
		VHost string `json:"vhost"`
		Type  string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiQueues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var streams []Stream
	for _, q := range apiQueues {
		if q.Type == "stream" {
			streams = append(streams, Stream{
				Name:         q.Name,
				ConnectionID: c.ID,
				VHost:        q.VHost,
			})
		}
	}

	return streams, nil
}

// ListStreams returns all streams for this connection
func (c *Connection) ListStreams(ctx context.Context) ([]Stream, error) {
	vhost := c.Config.VHost
	if vhost == "" {
		vhost = "/"
	}

	// Use RabbitMQ Management API to list queues (streams are queues with type="stream")
	apiURL := fmt.Sprintf("%s/api/queues/%s", c.Config.ManagementURL(), url.PathEscape(vhost))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query management API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("management API returned %d: %s", resp.StatusCode, string(body))
	}

	var apiQueues []struct {
		Name  string `json:"name"`
		VHost string `json:"vhost"`
		Type  string `json:"type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiQueues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter for streams only (queues with type="stream")
	var streams []Stream
	for _, q := range apiQueues {
		if q.Type == "stream" {
			streams = append(streams, Stream{
				Name:         q.Name,
				ConnectionID: c.ID,
				VHost:        q.VHost,
			})
		}
	}

	return streams, nil
}

// GetStreamStats returns statistics for a stream (using connection's default vhost)
func (c *Connection) GetStreamStats(ctx context.Context, streamName string) (*StreamStats, error) {
	vhost := c.Config.VHost
	if vhost == "" {
		vhost = "/"
	}
	return c.GetStreamStatsForVHost(ctx, vhost, streamName)
}

// GetStreamStatsForVHost returns statistics for a stream in a specific vhost
func (c *Connection) GetStreamStatsForVHost(ctx context.Context, vhost, streamName string) (*StreamStats, error) {
	// Query stream metadata (streams are queues with type="stream")
	apiURL := fmt.Sprintf("%s/api/queues/%s/%s", 
		c.Config.ManagementURL(), 
		url.PathEscape(vhost), 
		url.PathEscape(streamName))
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query management API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("management API returned %d: %s", resp.StatusCode, string(body))
	}

	var streamInfo struct {
		Name      string `json:"name"`
		Messages  int64  `json:"messages"`
		Backing   struct {
			Size int64 `json:"size"`
		} `json:"backing_queue_status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&streamInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Get first and last offset using stream protocol
	firstOffset, lastOffset, err := c.getStreamOffsetsForVHost(vhost, streamName)
	if err != nil {
		// If we can't get offsets, return what we have
		firstOffset = 0
		lastOffset = 0
	}

	return &StreamStats{
		Name:         streamName,
		MessageCount: streamInfo.Messages,
		Size:         streamInfo.Backing.Size,
		FirstOffset:  firstOffset,
		LastOffset:   lastOffset,
	}, nil
}

// getStreamOffsets retrieves the first and last offsets for a stream (using connection's default vhost)
func (c *Connection) getStreamOffsets(streamName string) (uint64, uint64, error) {
	// Query first offset
	firstOffset, err := c.Environment.QueryOffset(streamName, "first")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query first offset: %w", err)
	}

	// Query last offset
	lastOffset, err := c.Environment.QueryOffset(streamName, "last")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query last offset: %w", err)
	}

	return uint64(firstOffset), uint64(lastOffset), nil
}

// getStreamOffsetsForVHost retrieves the first and last offsets for a stream in a specific vhost
func (c *Connection) getStreamOffsetsForVHost(vhost, streamName string) (uint64, uint64, error) {
	// Create a new environment for this vhost
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(c.Config.Host).
			SetPort(c.Config.StreamPort).
			SetUser(c.Config.Username).
			SetPassword(c.Config.Password).
			SetVHost(vhost),
	)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create environment for vhost %s: %w", vhost, err)
	}
	defer env.Close()

	// Query first offset
	firstOffset, err := env.QueryOffset(streamName, "first")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query first offset: %w", err)
	}

	// Query last offset
	lastOffset, err := env.QueryOffset(streamName, "last")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query last offset: %w", err)
	}

	return uint64(firstOffset), uint64(lastOffset), nil
}

// ReadMessages reads messages from a stream starting at the given offset (using connection's default vhost)
func (c *Connection) ReadMessages(ctx context.Context, streamName string, offset uint64, limit int) (*MessageBatch, error) {
	vhost := c.Config.VHost
	if vhost == "" {
		vhost = "/"
	}
	return c.ReadMessagesFromVHost(ctx, vhost, streamName, offset, limit)
}

// ReadMessagesFromVHost reads messages from a stream in a specific vhost starting at the given offset
func (c *Connection) ReadMessagesFromVHost(ctx context.Context, vhost, streamName string, offset uint64, limit int) (*MessageBatch, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 500 {
		limit = 500
	}

	// Create a new environment for this vhost
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(c.Config.Host).
			SetPort(c.Config.StreamPort).
			SetUser(c.Config.Username).
			SetPassword(c.Config.Password).
			SetVHost(vhost),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment for vhost %s: %w", vhost, err)
	}
	defer env.Close()

	messages := make([]Message, 0, limit)
	mu := sync.Mutex{}
	done := make(chan bool)
	errChan := make(chan error, 1)

	consumer, err := env.NewConsumer(
		streamName,
		func(consumerContext stream.ConsumerContext, message *amqp.Message) {
			mu.Lock()
			defer mu.Unlock()

			if len(messages) >= limit {
				return
			}

			// Extract AMQP properties
			props := make(map[string]interface{})
			
			// AMQP Message Properties
			if message.Properties != nil {
				if message.Properties.MessageID != nil {
					props["message_id"] = message.Properties.MessageID
				}
				if message.Properties.CorrelationID != nil {
					props["correlation_id"] = message.Properties.CorrelationID
				}
				if message.Properties.ContentType != "" {
					props["content_type"] = message.Properties.ContentType
				}
				if message.Properties.ContentEncoding != "" {
					props["content_encoding"] = message.Properties.ContentEncoding
				}
				if message.Properties.ReplyTo != "" {
					props["reply_to"] = message.Properties.ReplyTo
				}
				if message.Properties.Subject != "" {
					props["subject"] = message.Properties.Subject
				}
				if message.Properties.To != "" {
					props["to"] = message.Properties.To
				}
				if len(message.Properties.UserID) > 0 {
					props["user_id"] = string(message.Properties.UserID)
				}
				if message.Properties.GroupID != "" {
					props["group_id"] = message.Properties.GroupID
				}
				if message.Properties.ReplyToGroupID != "" {
					props["reply_to_group_id"] = message.Properties.ReplyToGroupID
				}
				if message.Properties.GroupSequence != 0 {
					props["group_sequence"] = message.Properties.GroupSequence
				}
				if !message.Properties.CreationTime.IsZero() {
					props["creation_time"] = message.Properties.CreationTime
				}
				if !message.Properties.AbsoluteExpiryTime.IsZero() {
					props["absolute_expiry_time"] = message.Properties.AbsoluteExpiryTime
				}
			}

			// AMQP Header (message header fields)
			if message.Header != nil {
				if message.Header.Durable {
					props["durable"] = message.Header.Durable
				}
				if message.Header.Priority != 0 {
					props["priority"] = message.Header.Priority
				}
				if message.Header.TTL != 0 {
					props["ttl"] = message.Header.TTL
				}
				if message.Header.FirstAcquirer {
					props["first_acquirer"] = message.Header.FirstAcquirer
				}
				if message.Header.DeliveryCount != 0 {
					props["delivery_count"] = message.Header.DeliveryCount
				}
			}

			// Message Annotations (used by brokers and infrastructure)
			if message.Annotations != nil && len(message.Annotations) > 0 {
				annotations := make(map[string]interface{})
				for k, v := range message.Annotations {
					keyStr := fmt.Sprintf("%v", k)
					annotations[keyStr] = v
					
					// Extract routing key if present in annotations
					if keyStr == "x-routing-key" || keyStr == "routing-key" {
						props["routing_key"] = v
					}
				}
				props["message_annotations"] = annotations
			}

			// Delivery Annotations (used by delivery infrastructure)
			if message.DeliveryAnnotations != nil && len(message.DeliveryAnnotations) > 0 {
				deliveryAnnotations := make(map[string]interface{})
				for k, v := range message.DeliveryAnnotations {
					keyStr := fmt.Sprintf("%v", k)
					deliveryAnnotations[keyStr] = v
					
					// Extract routing key if present in delivery annotations
					if keyStr == "x-routing-key" || keyStr == "routing-key" {
						props["routing_key"] = v
					}
				}
				props["delivery_annotations"] = deliveryAnnotations
			}

			// Application Properties (custom key-value pairs set by the application)
			if message.ApplicationProperties != nil {
				appProps := make(map[string]interface{})
				for k, v := range message.ApplicationProperties {
					appProps[k] = v
				}
				if len(appProps) > 0 {
					props["application_properties"] = appProps
				}
			}

			// Footer (used for signatures, checksums, etc.)
			if message.Footer != nil && len(message.Footer) > 0 {
				footer := make(map[string]interface{})
				for k, v := range message.Footer {
					footer[fmt.Sprintf("%v", k)] = v
				}
				props["footer"] = footer
			}

			msg := Message{
				Offset:     uint64(consumerContext.Consumer.GetOffset()),
				Timestamp:  time.Now(),
				Data:       message.GetData(),
				Properties: props,
			}

			messages = append(messages, msg)

			if len(messages) >= limit {
				select {
				case done <- true:
				default:
				}
			}
		},
		stream.NewConsumerOptions().
			SetOffset(stream.OffsetSpecification{}.Offset(int64(offset))),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Wait for messages or timeout
	select {
	case <-done:
		// Got all messages
	case err := <-errChan:
		consumer.Close()
		return nil, err
	case <-time.After(5 * time.Second):
		// Timeout - return what we have
	case <-ctx.Done():
		consumer.Close()
		return nil, ctx.Err()
	}

	consumer.Close()

	mu.Lock()
	defer mu.Unlock()

	if len(messages) == 0 {
		return &MessageBatch{
			Messages:     []Message{},
			StartOffset:  offset,
			EndOffset:    offset,
			HasMore:      false,
		}, nil
	}

	return &MessageBatch{
		Messages:     messages,
		StartOffset:  messages[0].Offset,
		EndOffset:    messages[len(messages)-1].Offset,
		HasMore:      len(messages) == limit,
	}, nil
}

