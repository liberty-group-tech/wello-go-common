package kafka

import (
	"context"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/liberty-group-tech/wello-go-common/logging"
	"github.com/samber/lo"
)

const (
	defaultGroupID  = "default-group"
	defaultClientID = "default-client"
)

type Config struct {
	Brokers           []string
	ClientID          string
	GroupID           string
	Environment       string
	Topics            []string
	Partitions        int
	ReplicationFactor int
	Logger            logging.Logger
	Headers           []ckafka.Header
}

type configOptions struct {
	brokers           []string
	clientID          string
	groupID           string
	environment       string
	topics            []string
	partitions        int
	replicationFactor int
	logger            logging.Logger
	headers           []ckafka.Header
}

type Option func(*configOptions)

func WithBrokers(brokers []string) Option {
	return func(o *configOptions) {
		o.brokers = brokers
	}
}

func WithTopics(topics []string) Option {
	return func(o *configOptions) {
		o.topics = topics
	}
}

func WithPartitions(partitions int) Option {
	return func(o *configOptions) {
		o.partitions = partitions
	}
}

func WithReplicationFactor(replicationFactor int) Option {
	return func(o *configOptions) {
		o.replicationFactor = replicationFactor
	}
}

func WithGroupID(groupID string) Option {
	return func(o *configOptions) {
		o.groupID = groupID
	}
}

func WithClientID(clientID string) Option {
	return func(o *configOptions) {
		o.clientID = clientID
	}
}

func WithEnvironment(environment string) Option {
	return func(o *configOptions) {
		o.environment = environment
	}
}

func WithLogger(logger logging.Logger) Option {
	return func(o *configOptions) {
		o.logger = logger
	}
}

func NewConfig(opts ...Option) *Config {
	options := &configOptions{
		clientID:          defaultClientID,
		groupID:           defaultGroupID,
		partitions:        3,
		replicationFactor: 1,
		environment:       "",
		headers: []ckafka.Header{
			{Key: "Content-Type", Value: []byte("application/json")},
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	logger := options.logger
	if logger == nil {
		logger = &logging.NoOpLogger{}
	}

	return &Config{
		Brokers:           options.brokers,
		ClientID:          options.clientID,
		GroupID:           options.groupID,
		Environment:       options.environment,
		Topics:            options.topics,
		Partitions:        options.partitions,
		ReplicationFactor: options.replicationFactor,
		Logger:            logger,
	}
}

func WithHeaders(headers []ckafka.Header) Option {
	return func(o *configOptions) {
		o.headers = append(o.headers, headers...)
	}
}

func ensureTopics(cfg *Config, producer *ckafka.Producer, consumer *ckafka.Consumer) error {

	var admin *ckafka.AdminClient
	var err error
	if producer != nil {
		admin, err = ckafka.NewAdminClientFromProducer(producer)
	} else {
		admin, err = ckafka.NewAdminClientFromConsumer(consumer)
	}
	if err != nil {
		return err
	}
	defer admin.Close()

	if len(cfg.Topics) == 0 {
		return nil
	}

	topics := lo.Map(cfg.Topics, func(topic string, _ int) ckafka.TopicSpecification {
		return ckafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     cfg.Partitions,
			ReplicationFactor: cfg.ReplicationFactor,
		}
	})

	_, err = admin.CreateTopics(context.TODO(), topics, nil)
	if err != nil && !strings.Contains(err.Error(), "Topic already exists") {
		return err
	}
	return nil
}
