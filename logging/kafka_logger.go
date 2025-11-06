package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap/zapcore"
)

type KafkaCore struct {
	producer *kafka.Producer
	topic    string
	encoder  zapcore.Encoder
	level    zapcore.Level
	appName  string
}

func NewKafkaCore(brokers []string, topic string, encoder zapcore.Encoder, level zapcore.Level, appName string) (*KafkaCore, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"client.id":         "wello-go-common-logger",
		"acks":              "1",
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaCore{
		producer: producer,
		topic:    topic,
		encoder:  encoder,
		level:    level,
		appName:  appName,
	}, nil
}

func (kc *KafkaCore) Enabled(level zapcore.Level) bool {
	return level >= kc.level
}

func (kc *KafkaCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *kc
	clone.encoder = kc.encoder.Clone()
	for _, field := range fields {
		field.AddTo(clone.encoder)
	}
	return &clone
}

func (kc *KafkaCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if kc.Enabled(ent.Level) {
		return ce.AddCore(ent, kc)
	}
	return ce
}

func (kc *KafkaCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := kc.encoder.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	defer buf.Free()

	logMessage := map[string]interface{}{
		"@timestamp": ent.Time.Format(time.RFC3339Nano),
		"appName":    kc.appName,
		"level":      ent.Level.String(),
		"message":    ent.Message,
		"caller":     ent.Caller.String(),
	}

	// Extract additional fields using a custom ObjectEncoder
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	// Merge extracted fields into logMessage
	for k, v := range enc.Fields {
		logMessage[k] = v
	}

	jsonData, err := json.Marshal(logMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kc.topic,
			Partition: kafka.PartitionAny,
		},
		Value: jsonData,
		Key:   []byte(fmt.Sprintf("%s-%s", ent.Level.String(), ent.Time.Format(time.RFC3339Nano))),
	}

	err = kc.producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to produce kafka message: %w", err)
	}
	return nil
}

func (kc *KafkaCore) Sync() error {
	kc.producer.Flush(5000)
	return nil
}

func (kc *KafkaCore) Close() error {
	kc.producer.Close()
	return nil
}
