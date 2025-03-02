package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/your-username/slido-clone/user-service/config"
)

// EventType represents the type of event
type EventType string

// Event types
const (
	// User events
	UserCreated     EventType = "user.created"
	UserUpdated     EventType = "user.updated"
	UserDeleted     EventType = "user.deleted"
	UserActivated   EventType = "user.activated"
	UserDeactivated EventType = "user.deactivated"

	// Team events
	TeamCreated       EventType = "team.created"
	TeamUpdated       EventType = "team.updated"
	TeamDeleted       EventType = "team.deleted"
	TeamMemberAdded   EventType = "team.member.added"
	TeamMemberUpdated EventType = "team.member.updated"
	TeamMemberRemoved EventType = "team.member.removed"

	// Organization events
	OrganizationCreated       EventType = "organization.created"
	OrganizationUpdated       EventType = "organization.updated"
	OrganizationDeleted       EventType = "organization.deleted"
	OrganizationMemberAdded   EventType = "organization.member.added"
	OrganizationMemberUpdated EventType = "organization.member.updated"
	OrganizationMemberRemoved EventType = "organization.member.removed"
)

// Event represents a Kafka event
type Event struct {
	ID            string      `json:"id"`
	Type          EventType   `json:"type"`
	Source        string      `json:"source"`
	Subject       string      `json:"subject,omitempty"`
	Time          time.Time   `json:"time"`
	Data          interface{} `json:"data"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// Producer is a Kafka producer
type Producer struct {
	producer *kafka.Producer
	config   *config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	// Create Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers[0], // Use the first broker
		"client.id":         cfg.ClientID,
		"acks":              "all", // Wait for all replicas to acknowledge
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create Kafka producer")
		return nil, err
	}

	// Start a goroutine to handle delivery reports
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Error().
						Err(ev.TopicPartition.Error).
						Str("topic", *ev.TopicPartition.Topic).
						Int32("partition", ev.TopicPartition.Partition).
						Msg("Failed to deliver message")
				} else {
					log.Debug().
						Str("topic", *ev.TopicPartition.Topic).
						Int32("partition", ev.TopicPartition.Partition).
						Int64("offset", ev.TopicPartition.Offset).
						Msg("Message delivered")
				}
			case kafka.Error:
				log.Error().
					Str("code", ev.Code().String()).
					Msg("Kafka error")
			}
		}
	}()

	log.Info().Msg("Kafka producer created")

	return &Producer{
		producer: p,
		config:   cfg,
	}, nil
}

// Close closes the Kafka producer
func (p *Producer) Close() {
	p.producer.Flush(15 * 1000) // Wait up to 15s for messages to be delivered
	p.producer.Close()
	log.Info().Msg("Kafka producer closed")
}

// PublishUserEvent publishes a user event
func (p *Producer) PublishUserEvent(eventType EventType, data interface{}, subject string, correlationID string) error {
	return p.publish(p.config.Topics.UserEvents, eventType, data, subject, correlationID)
}

// PublishTeamEvent publishes a team event
func (p *Producer) PublishTeamEvent(eventType EventType, data interface{}, subject string, correlationID string) error {
	return p.publish(p.config.Topics.TeamEvents, eventType, data, subject, correlationID)
}

// publish publishes an event to Kafka
func (p *Producer) publish(topic string, eventType EventType, data interface{}, subject string, correlationID string) error {
	// Create event
	event := Event{
		ID:            uuid.New().String(),
		Type:          eventType,
		Source:        "user-service",
		Subject:       subject,
		Time:          time.Now(),
		Data:          data,
		CorrelationID: correlationID,
	}

	// Serialize event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal event")
		return err
	}

	// Create message key
	var key string
	if subject != "" {
		key = subject
	} else {
		key = string(eventType)
	}

	// Create Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: eventBytes,
		Headers: []kafka.Header{
			{
				Key:   "event-type",
				Value: []byte(eventType),
			},
			{
				Key:   "source",
				Value: []byte("user-service"),
			},
			{
				Key:   "id",
				Value: []byte(event.ID),
			},
			{
				Key:   "time",
				Value: []byte(event.Time.Format(time.RFC3339)),
			},
		},
	}

	// Add correlation ID header if provided
	if correlationID != "" {
		message.Headers = append(message.Headers, kafka.Header{
			Key:   "correlation-id",
			Value: []byte(correlationID),
		})
	}

	// Produce message
	if err := p.producer.Produce(message, nil); err != nil {
		log.Error().
			Err(err).
			Str("topic", topic).
			Str("event_type", string(eventType)).
			Msg("Failed to produce message")
		return fmt.Errorf("failed to produce message: %w", err)
	}

	log.Debug().
		Str("topic", topic).
		Str("event_type", string(eventType)).
		Str("event_id", event.ID).
		Msg("Message produced")

	return nil
}
