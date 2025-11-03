package kafka

// Config 配置接口
type Config interface {
	GetBrokers() []string
	GetClientID() string
	GetEnvironment() string
}

// SimpleConfig 简单的配置实现
type SimpleConfig struct {
	Brokers     []string
	ClientID    string
	Environment string
}

func (c *SimpleConfig) GetBrokers() []string {
	return c.Brokers
}

func (c *SimpleConfig) GetClientID() string {
	return c.ClientID
}

func (c *SimpleConfig) GetEnvironment() string {
	return c.Environment
}
