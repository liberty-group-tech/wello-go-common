package kafka

import (
	"strings"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/liberty-group-tech/wello-go-common/logging"
)

// ConsumerPool 消费者池
type ConsumerPool struct {
	mu      sync.Mutex
	cfg     Config
	groupID string
	topic   string
	pool    chan *ckafka.Consumer
}

var (
	poolMap = make(map[string]*ConsumerPool)
	mapMu   sync.Mutex
)

func poolKey(groupID string, topic string) string {
	return groupID + "::" + topic
}

// ConsumerManager 消费者管理器
type ConsumerManager struct {
	cfg     Config
	logger  logging.Logger
	groupID string
}

// NewConsumerManager 创建消费者管理器
func NewConsumerManager(cfg Config, logger logging.Logger, groupID string) *ConsumerManager {
	if logger == nil {
		logger = &logging.NoOpLogger{}
	}
	return &ConsumerManager{
		cfg:     cfg,
		logger:  logger,
		groupID: groupID,
	}
}

// GetPool 获取指定主题的消费者池
func (cm *ConsumerManager) GetPool(topic string) *ConsumerPool {
	key := poolKey(cm.groupID, topic)

	mapMu.Lock()
	defer mapMu.Unlock()

	if pool, ok := poolMap[key]; ok {
		return pool
	}

	cp := &ConsumerPool{
		cfg:     cm.cfg,
		groupID: cm.groupID,
		topic:   topic,
		pool:    make(chan *ckafka.Consumer, 1000),
	}
	poolMap[key] = cp
	return cp
}

// Borrow 从池中借用消费者，返回消费者和归还函数
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

// Return 归还消费者到池中
func (cp *ConsumerPool) Return(consumer *ckafka.Consumer) {
	cp.pool <- consumer
}

func (cp *ConsumerPool) createAndPut() *ckafka.Consumer {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(cp.cfg.GetBrokers(), ","),
		"group.id":          cp.groupID,
		"client.id":         cp.cfg.GetClientID(),
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	if err := consumer.Subscribe(cp.topic, nil); err != nil {
		panic(err)
	}

	cp.pool <- consumer
	return <-cp.pool
}
