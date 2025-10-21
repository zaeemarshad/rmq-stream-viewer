package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

type Message struct {
	Timestamp string `json:"timestamp"`
	Counter   int    `json:"counter"`
}

func main() {
	streamName := flag.String("stream", "test-stream", "Stream name to publish to")
	host := flag.String("host", "localhost", "RabbitMQ host")
	port := flag.Int("port", 5552, "RabbitMQ stream port")
	user := flag.String("user", "root", "RabbitMQ username")
	password := flag.String("password", "magic", "RabbitMQ password")
	vhost := flag.String("vhost", "/", "RabbitMQ vhost")
	rate := flag.Int("rate", 1, "Messages per second")
	count := flag.Int("count", 0, "Total messages to send (0 = infinite)")
	flag.Parse()

	log.Printf("Starting publisher for stream '%s'", *streamName)
	log.Printf("Connecting to %s:%d", *host, *port)
	log.Printf("Publishing at %d msg/sec", *rate)

	// Create environment
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(*host).
			SetPort(*port).
			SetUser(*user).
			SetPassword(*password).
			SetVHost(*vhost),
	)
	if err != nil {
		log.Fatalf("Failed to create environment: %v", err)
	}
	defer env.Close()

	// Check if stream exists, create if it doesn't
	exists, err := env.StreamExists(*streamName)
	if err != nil {
		log.Fatalf("Failed to check stream existence: %v", err)
	}

	if !exists {
		log.Printf("Creating stream '%s'", *streamName)
		err = env.DeclareStream(*streamName,
			&stream.StreamOptions{
				MaxLengthBytes: stream.ByteCapacity{}.GB(2),
			},
		)
		if err != nil {
			log.Fatalf("Failed to create stream: %v", err)
		}
	}

	// Create producer
	producer, err := env.NewProducer(*streamName, nil)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	log.Println("Publisher started successfully")

	// Publish messages
	counter := 0
	ticker := time.NewTicker(time.Second / time.Duration(*rate))
	defer ticker.Stop()

	for range ticker.C {
		counter++

		// Create message payload
		msg := Message{
			Timestamp: time.Now().Format(time.RFC3339),
			Counter:   counter,
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		// Create AMQP message with properties
		messageID := uuid.New().String()
		contentType := "application/json"
		timestamp := time.Now()
		subject := fmt.Sprintf("test-message-%d", counter)

		amqpMsg := amqp.NewMessage(payload)
		amqpMsg.Properties = &amqp.MessageProperties{
			MessageID:    &messageID,
			ContentType:  contentType,
			CreationTime: timestamp,
			Subject:      subject,
		}
		amqpMsg.ApplicationProperties = map[string]interface{}{
			"app_id":      "test-publisher",
			"counter":     counter,
			"environment": "test",
		}

		// Send message
		err = producer.Send(amqpMsg)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
			continue
		}

		log.Printf("Published message %d (ID: %s)", counter, messageID)

		// Check if we've reached the count limit
		if *count > 0 && counter >= *count {
			log.Printf("Reached message count limit (%d), exiting", *count)
			break
		}
	}

	log.Println("Publisher stopped")
}

