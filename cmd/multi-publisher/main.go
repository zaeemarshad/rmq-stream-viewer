package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

type Message struct {
	VHost     string `json:"vhost"`
	Stream    string `json:"stream"`
	Timestamp string `json:"timestamp"`
	Counter   int    `json:"counter"`
	Data      string `json:"data"`
}

type StreamPublisher struct {
	vhost    string
	stream   string
	env      *stream.Environment
	producer *stream.Producer
	counter  int
}

func main() {
	host := flag.String("host", "localhost", "RabbitMQ host")
	port := flag.Int("port", 5552, "RabbitMQ stream port")
	user := flag.String("user", "root", "RabbitMQ username")
	password := flag.String("password", "magic", "RabbitMQ password")
	rate := flag.Int("rate", 2, "Messages per second per stream")
	count := flag.Int("count", 100, "Total messages to send per stream")
	flag.Parse()

	log.Printf("Starting multi-vhost publisher")
	log.Printf("Connecting to %s:%d", *host, *port)
	log.Printf("Publishing at %d msg/sec per stream", *rate)

	// Define vhosts and streams to create
	config := map[string][]string{
		"/":            {"default-stream-1", "default-stream-2"},
		"test-vhost-1": {"orders", "payments", "notifications"},
		"test-vhost-2": {"logs", "metrics", "events"},
		"test-vhost-3": {"analytics", "reports"},
	}

	var publishers []*StreamPublisher
	var wg sync.WaitGroup

	// Create publishers for each vhost/stream combination
	for vhost, streams := range config {
		for _, streamName := range streams {
			pub, err := createPublisher(*host, *port, *user, *password, vhost, streamName)
			if err != nil {
				log.Printf("Failed to create publisher for %s/%s: %v", vhost, streamName, err)
				continue
			}
			publishers = append(publishers, pub)
			log.Printf("Created publisher for vhost='%s' stream='%s'", vhost, streamName)
		}
	}

	log.Printf("Successfully created %d publishers", len(publishers))

	// Start publishing to all streams concurrently
	for _, pub := range publishers {
		wg.Add(1)
		go func(p *StreamPublisher) {
			defer wg.Done()
			publishToStream(p, *rate, *count)
		}(pub)
	}

	// Wait for all publishers to finish
	wg.Wait()

	// Close all publishers
	for _, pub := range publishers {
		if pub.producer != nil {
			pub.producer.Close()
		}
		if pub.env != nil {
			pub.env.Close()
		}
	}

	log.Println("All publishers stopped")
}

func createPublisher(host string, port int, user, password, vhost, streamName string) (*StreamPublisher, error) {
	// Create environment
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(host).
			SetPort(port).
			SetUser(user).
			SetPassword(password).
			SetVHost(vhost),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Check if stream exists, create if it doesn't
	exists, err := env.StreamExists(streamName)
	if err != nil {
		env.Close()
		return nil, fmt.Errorf("failed to check stream existence: %w", err)
	}

	if !exists {
		log.Printf("Creating stream '%s' in vhost '%s'", streamName, vhost)
		err = env.DeclareStream(streamName,
			&stream.StreamOptions{
				MaxLengthBytes: stream.ByteCapacity{}.MB(100),
			},
		)
		if err != nil {
			env.Close()
			return nil, fmt.Errorf("failed to create stream: %w", err)
		}
	}

	// Create producer
	producer, err := env.NewProducer(streamName, nil)
	if err != nil {
		env.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &StreamPublisher{
		vhost:    vhost,
		stream:   streamName,
		env:      env,
		producer: producer,
		counter:  0,
	}, nil
}

func publishToStream(pub *StreamPublisher, rate, count int) {
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for range ticker.C {
		pub.counter++

		// Create message payload with some sample data
		msg := Message{
			VHost:     pub.vhost,
			Stream:    pub.stream,
			Timestamp: time.Now().Format(time.RFC3339),
			Counter:   pub.counter,
			Data:      fmt.Sprintf("Sample data for message %d in %s/%s", pub.counter, pub.vhost, pub.stream),
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[%s/%s] Failed to marshal message: %v", pub.vhost, pub.stream, err)
			continue
		}

		// Create AMQP message with properties
		messageID := uuid.New().String()
		contentType := "application/json"
		timestamp := time.Now()
		subject := fmt.Sprintf("%s-msg-%d", pub.stream, pub.counter)

		amqpMsg := amqp.NewMessage(payload)
		amqpMsg.Properties = &amqp.MessageProperties{
			MessageID:    &messageID,
			ContentType:  contentType,
			CreationTime: timestamp,
			Subject:      subject,
		}
		amqpMsg.ApplicationProperties = map[string]interface{}{
			"app_id":      "multi-publisher",
			"counter":     pub.counter,
			"vhost":       pub.vhost,
			"stream":      pub.stream,
			"environment": "test",
		}

		// Add routing key to annotations
		amqpMsg.Annotations = map[interface{}]interface{}{
			"x-routing-key": fmt.Sprintf("%s.%s.%d", pub.vhost, pub.stream, pub.counter),
		}

		// Send message
		err = pub.producer.Send(amqpMsg)
		if err != nil {
			log.Printf("[%s/%s] Failed to send message: %v", pub.vhost, pub.stream, err)
			continue
		}

		if pub.counter%10 == 0 {
			log.Printf("[%s/%s] Published %d messages", pub.vhost, pub.stream, pub.counter)
		}

		// Check if we've reached the count limit
		if count > 0 && pub.counter >= count {
			log.Printf("[%s/%s] Reached message count limit (%d), stopping", pub.vhost, pub.stream, count)
			break
		}
	}
}

