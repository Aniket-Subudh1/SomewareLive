package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/config"
)

// Handler is a function that handles a Kafka message
type Handler func(ctx context.Context, event Event) error

// Consumer is a Kafka consumer
type Consumer struct {
	consumer      *kafka.Consumer
	config        *config.KafkaConfig
	handlers      map[string]map[EventType]Handler
	shutdownCh    chan struct{}
	subscriptions []string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.KafkaConfig) (*Consumer, error) {
	// Create Kafka consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.Brokers[0], // Use the first broker
		"group.id":           cfg.GroupID,
		"client.id":          cfg.ClientID,
		"auto.offset.reset":  cfg.AutoOffsetReset,
		"enable.auto.commit": true,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kafka consumer")
		return nil, err
	}

	log.Info().Msg("Kafka consumer created")

	return &Consumer{
		consumer:      c,
		config:        cfg,
		handlers:      make(map[string]map[EventType]Handler),
		shutdownCh:    make(chan struct{}),
		subscriptions: make([]string, 0),
	}, nil
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(topic string, eventType EventType, handler Handler) {
	// Add topic to subscriptions if it's not already there
	alreadySubscribed := false
	for _, t := range c.subscriptions {
		if t == topic {
			alreadySubscribed = true
			break
		}
	}
	if !alreadySubscribed {
		c.subscriptions = append(c.subscriptions, topic)
	}

	// Initialize topic handlers map if needed
	if _, ok := c.handlers[topic]; !ok {
		c.handlers[topic] = make(map[EventType]Handler)
	}

	// Register handler
	c.handlers[topic][eventType] = handler
	log.Info().Str("topic", topic).Str("event_type", string(eventType)).Msg("Registered event handler")
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	// Make sure we have subscriptions
	if len(c.subscriptions) == 0 {
		log.Warn().Msg("No subscriptions registered, consumer won't start")
		return errors.New("no subscriptions registered")
	}

	// Subscribe to topics
	if err := c.consumer.SubscribeTopics(c.subscriptions, nil); err != nil {
		log.Error().Err(err).Strs("topics", c.subscriptions).Msg("Failed to subscribe to topics")
		return err
	}

	log.Info().Strs("topics", c.subscriptions).Msg("Subscribed to topics")

	// Start consumer loop
	go c.consume(ctx)

	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() {
	close(c.shutdownCh)
	c.consumer.Close()
	log.Info().Msg("Kafka consumer closed")
}

// consume consumes messages from Kafka
func (c *Consumer) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping consumer")
			return
		case <-c.shutdownCh:
			log.Info().Msg("Shutdown signal received, stopping consumer")
			return
		default:
			// Poll for messages
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Ignore timeout errors
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Error().Err(err).Msg("Error reading message")
				continue
			}

			// Process message
			if err := c.processMessage(ctx, msg); err != nil {
				log.Error().Err(err).Msg("Error processing message")
			}
		}
	}
}

// processMessage processes a Kafka message
func (c *Consumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic

	// Parse event
	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Error().
			Err(err).
			Str("topic", topic).
			Bytes("value", msg.Value).
			Msg("Failed to unmarshal event")
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Get event type from header if possible
	var eventType EventType
	for _, header := range msg.Headers {
		if header.Key == "event-type" {
			eventType = EventType(header.Value)
			break
		}
	}

	// If event type not found in header, use from event
	if eventType == "" {
		eventType = event.Type
	}

	// Check if we have a handler for this topic and event type
	topicHandlers, ok := c.handlers[topic]
	if !ok {
		log.Debug().
			Str("topic", topic).
			Str("event_type", string(eventType)).
			Msg("No handlers for topic")
		return nil
	}

	handler, ok := topicHandlers[eventType]
	if !ok {
		handler, ok = topicHandlers["*"] // Check for wildcard handler
		if !ok {
			log.Debug().
				Str("topic", topic).
				Str("event_type", string(eventType)).
				Msg("No handler for event type")
			return nil
		}
	}

	// Get correlation ID if available
	var correlationID string
	for _, header := range msg.Headers {
		if header.Key == "correlation-id" {
			correlationID = string(header.Value)
			break
		}
	}

	// If correlation ID not found in header, use from event
	if correlationID == "" {
		correlationID = event.CorrelationID
	}

	// Create a context with correlation ID
	handlerCtx := ctx
	if correlationID != "" {
		handlerCtx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	// Handle event with logging
	log.Debug().
		Str("topic", topic).
		Str("event_type", string(eventType)).
		Str("event_id", event.ID).
		Str("correlation_id", correlationID).
		Msg("Processing event")

	startTime := time.Now()
	err := handler(handlerCtx, event)
	duration := time.Since(startTime)

	if err != nil {
		log.Error().
			Err(err).
			Str("topic", topic).
			Str("event_type", string(eventType)).
			Str("event_id", event.ID).
			Str("correlation_id", correlationID).
			Dur("duration", duration).
			Msg("Error handling event")
		return fmt.Errorf("error handling event: %w", err)
	}

	log.Debug().
		Str("topic", topic).
		Str("event_type", string(eventType)).
		Str("event_id", event.ID).
		Str("correlation_id", correlationID).
		Dur("duration", duration).
		Msg("Event processed successfully")

	return nil
}
