package kafka

import (
	"strings"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/liberty-group-tech/wello-go-common/logging"
)

type ConsumerPool struct {
	mu      sync.Mutex
	cfg     *Config
	groupID string
	topic   string
	pool    chan *ckafka.Consumer
	once    sync.Once
}

var (
	poolMap = make(map[string]*ConsumerPool)
	mapMu   sync.Mutex
)

func poolKey(groupID string, topic string) string {
	return groupID + "::" + topic
}

type ConsumerManager struct {
	cfg    *Config
	logger logging.Logger
}

func NewConsumerManager(opts ...Option) *ConsumerManager {

	cfg := newConfig(opts...)

	return &ConsumerManager{
		cfg:    cfg,
		logger: cfg.Logger,
	}
}

func (cm *ConsumerManager) GetPool(topic string) *ConsumerPool {
	key := poolKey(cm.cfg.GroupID, topic)

	mapMu.Lock()
	defer mapMu.Unlock()

	if pool, ok := poolMap[key]; ok {
		return pool
	}

	cp := &ConsumerPool{
		cfg:     cm.cfg,
		groupID: cm.cfg.GroupID,
		topic:   topic,
		pool:    make(chan *ckafka.Consumer, 1000),
	}
	poolMap[key] = cp
	return cp
}

func (cp *ConsumerPool) Borrow() (*ckafka.Consumer, func()) {
	var consumer *ckafka.Consumer
	var returnFunc func()
	select {
	case consumer = <-cp.pool:
		returnFunc = func() {
			cp.Return(consumer)
		}
	default:
		consumer = cp.createAndPut()
		returnFunc = func() {
			cp.Return(consumer)
		}
	}
	return consumer, returnFunc
}

func (cp *ConsumerPool) Return(consumer *ckafka.Consumer) {
	cp.pool <- consumer
}

func (cp *ConsumerPool) createAndPut() *ckafka.Consumer {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(cp.cfg.Brokers, ","),
		"group.id":          cp.groupID,
		"client.id":         cp.cfg.ClientID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	cp.once.Do(func() {
		if err := ensureTopics(cp.cfg, nil, consumer); err != nil {
			cp.cfg.Logger.Errorf("Failed to ensure topics: %v", err)
		}
	})

	if err := consumer.Subscribe(cp.topic, nil); err != nil {
		panic(err)
	}

	cp.pool <- consumer
	return <-cp.pool
}
