package redis

import (
	"context"
	"time"

	"github.com/liberty-group-tech/wello-go-common/helper"
	goredis "github.com/redis/go-redis/v9"
)

const lockPrefix = "redis-lock"

type ClientOptions = goredis.ClusterOptions

// Client wraps a redis cluster client with lazy loading support.
type Client struct {
	loader         *helper.Loader[*goredis.ClusterClient]
	mode           helper.LoadMode
	clusterOptions *goredis.ClusterOptions
}

type Option func(*Client)

func WithMode(mode helper.LoadMode) Option {
	return func(c *Client) {
		c.mode = mode
	}
}

func WithClusterOptions(opt *goredis.ClusterOptions) Option {
	return func(c *Client) {
		c.clusterOptions = opt
	}
}

// NewClient creates a redis client loader which connects to the provided hosts.
// The underlying client is created lazily on the first access.
func NewClient(opts ...Option) (*Client, error) {

	options := &Client{
		mode:           helper.Lazy,
		clusterOptions: &goredis.ClusterOptions{},
	}

	for _, opt := range opts {
		opt(options)
	}

	loader := helper.NewLoader(func() (*goredis.ClusterClient, error) {
		client := goredis.NewClusterClient(options.clusterOptions)
		if err := client.Ping(context.Background()).Err(); err != nil {
			return nil, err
		}
		return client, nil
	}, helper.WithMode(options.mode))

	return &Client{loader: loader}, nil
}

// Cluster returns the underlying redis cluster client, creating it if necessary.
func (c *Client) Cluster() (*goredis.ClusterClient, error) {
	return c.loader.Get()
}

// MustCluster returns the redis cluster client and panics if creation fails.
func (c *Client) MustCluster() *goredis.ClusterClient {
	return c.loader.MustGet()
}

// Close closes the underlying redis client when it has been created.
func (c *Client) Close() error {
	client, err := c.Cluster()
	if err != nil {
		return err
	}
	return client.Close()
}

// Set stores a value at the given key with expiration.
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	client, err := c.Cluster()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves the value of the given key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	client, err := c.Cluster()
	if err != nil {
		return "", err
	}
	return client.Get(ctx, key).Result()
}

// Del removes the given key.
func (c *Client) Del(ctx context.Context, key string) error {
	client, err := c.Cluster()
	if err != nil {
		return err
	}
	return client.Del(ctx, key).Err()
}

// Lock represents a lightweight redis distributed lock.
type Lock struct {
	Key        string
	token      string
	expiration time.Duration
	client     *Client
}

// AcquireLock tries to acquire a lock for the provided key.
// Returns nil if the lock is already held by another process.
func (c *Client) AcquireLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	client, err := c.Cluster()
	if err != nil {
		return nil, err
	}

	token := helper.GenerateID(lockPrefix)
	ok, err := client.SetNX(ctx, key, token, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &Lock{
		Key:        key,
		token:      token,
		expiration: expiration,
		client:     c,
	}, nil
}

// Refresh extends the lock expiration.
func (l *Lock) Refresh(ctx context.Context, expiration time.Duration) error {
	client, err := l.client.Cluster()
	if err != nil {
		return err
	}

	script := `if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("PEXPIRE", KEYS[1], ARGV[2])
    else
        return 0
    end`

	return client.Eval(ctx, script, []string{l.Key}, l.token, int(expiration/time.Millisecond)).Err()
}

// Release releases the lock if still owned by the caller.
func (l *Lock) Release(ctx context.Context) error {
	client, err := l.client.Cluster()
	if err != nil {
		return err
	}

	script := `if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end`

	return client.Eval(ctx, script, []string{l.Key}, l.token).Err()
}
