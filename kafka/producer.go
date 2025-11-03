package kafka

import (
	"context"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/liberty-group-tech/wello-go-common/logging"
)

// Producer Kafka 生产者
type Producer struct {
	producer *ckafka.Producer
	cfg      Config
	logger   logging.Logger
}

// NewProducer 创建生产者
func NewProducer(cfg Config, logger logging.Logger) (*Producer, error) {
	if logger == nil {
		logger = &logging.NoOpLogger{}
	}

	producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.GetBrokers(), ","),
		"client.id":         cfg.GetClientID(),
	})
	if err != nil {
		logger.Errorf("Failed to create kafka producer: %v", err)
		return nil, err
	}

	p := &Producer{
		producer: producer,
		cfg:      cfg,
		logger:   logger,
	}

	return p, nil
}

// SendMessage 发送消息
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

// EnsureTopics 确保主题存在
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

// Close 关闭生产者
func (p *Producer) Close() {
	p.producer.Close()
}

func (p *Producer) mergeHeaders(headers []ckafka.Header) []ckafka.Header {
	defaultHeaders := []ckafka.Header{
		{Key: "Content-Type", Value: []byte("application/json")},
	}
	if env := p.cfg.GetEnvironment(); env != "" {
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
