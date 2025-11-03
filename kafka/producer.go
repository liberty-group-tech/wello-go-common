package kafka

import (
	"context"
	"fmt"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/liberty-group-tech/wello-go-common/logging"
)

type Producer struct {
	producer *ckafka.Producer
	cfg      *Config
	logger   logging.Logger
}

func NewProducer(opts ...Option) (*Producer, error) {
	cfg := NewConfig(opts...)

	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
		"client.id":         cfg.ClientID,
	})
	if err != nil {
		cfg.Logger.Errorf("Failed to create kafka producer: %v", err)
		return nil, err
	}

	if err := ensureTopics(cfg, producer, nil); err != nil {
		cfg.Logger.Errorf("Failed to ensure topics: %v", err)
		// Note: Topic creation failure is logged but doesn't prevent producer creation
	}

	p := &Producer{
		producer: producer,
		cfg:      cfg,
		logger:   cfg.Logger,
	}

	return p, nil
}

func (p *Producer) SendMessage(topic string, message []byte, headers []ckafka.Header) error {
	finalHeaders := p.mergeHeaders(headers)
	return p.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: ckafka.PartitionAny,
		},
		Value:   message,
		Headers: finalHeaders,
	}, nil)
}

func (p *Producer) EnsureTopics(topics []string, partitions int, replicationFactor int) error {
	admin, err := ckafka.NewAdminClientFromProducer(p.producer)
	if err != nil {
		return err
	}
	defer admin.Close()

	specs := make([]ckafka.TopicSpecification, len(topics))
	for i, topic := range topics {
		specs[i] = ckafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		}
	}

	_, err = admin.CreateTopics(context.TODO(), specs, nil)
	if err != nil && !strings.Contains(err.Error(), "Topic already exists") {
		return err
	}
	return nil
}

func (p *Producer) Close() {
	p.producer.Close()
}

func (p *Producer) mergeHeaders(headers []ckafka.Header) []ckafka.Header {
	defaultHeaders := p.cfg.Headers
	if env := p.cfg.Environment; env != "" {
		defaultHeaders = append(defaultHeaders, ckafka.Header{
			Key:   "env",
			Value: []byte(env),
		})
	}
	if len(headers) == 0 {
		return defaultHeaders
	}
	return append(defaultHeaders, headers...)
}
