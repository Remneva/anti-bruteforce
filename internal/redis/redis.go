package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Client struct {
	rdb    *redis.Client
	l      *zap.Logger
	expiry time.Duration
}

type InterfaceRedis interface {
	RdbConnect(ctx context.Context, address string, password string) (*Client, error)
	GettingCount(ctx context.Context, key string) (int64, error)
	CleanByKey(ctx context.Context, key string) error
}

func NewClient(l *zap.Logger, expiry time.Duration) *Client {
	c := &Client{
		l:      l,
		expiry: expiry * time.Second,
	}
	return c
}
func (c *Client) RdbConnect(ctx context.Context, address string, password string) (*Client, error) {
	client := redis.NewClient(&redis.Options{ //nolint:exhaustivestruct
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	c.l.Info("getting Redis connection:", zap.String("response", pong))
	c.rdb = client
	return c, nil
}

func (c *Client) Close(ctx context.Context) error {
	err := c.rdb.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GettingCount(ctx context.Context, key string) (int64, error) {
	// count++
	incr := c.rdb.Incr(ctx, key)
	// count-- after timeout
	go c.decrement(ctx, key)
	count, err := incr.Result()
	c.l.Info("increment", zap.String("key", key), zap.Int64("incr", count))
	if err != nil {
		c.l.Error("getting increment error", zap.String("key", key))
		return 0, fmt.Errorf("getting increment error: %w", err)
	}
	// устанавливаем срок действия бакета, чтобы невостребованные ключи не накапливались
	if err = c.setTimeoutForKey(ctx, key); err != nil {
		c.l.Error("TTL setting error", zap.String("key", key))
		return 0, fmt.Errorf("TTL setting error: %w", err)
	}
	return count, nil
}

// устанавливаем срок действия инкремента 60 сек - по истечении минуты счетчик уменьшается на 1 ед.
func (c *Client) decrement(ctx context.Context, key string) {
	time.Sleep(c.expiry)
	incr := c.rdb.Decr(ctx, key)
	count, _ := incr.Result()
	c.l.Info("decrement", zap.String("key", key), zap.Int64("currently count", count))
}

func (c *Client) setTimeoutForKey(ctx context.Context, key string) error {
	// проверяем наличие ключа в базе
	alreadyExist, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		c.l.Error("getting key error", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("error while getting key: %w", err)
	}
	if alreadyExist == "1" { // "1" - ключ новый
		// устанавливаем срок действия ключа
		_, err = c.rdb.Expire(ctx, key, 1*time.Hour).Result()
		if err != nil {
			c.l.Error("setting TTL error", zap.String("key", key))
			return fmt.Errorf("error while setting TTL: %w", err)
		}
		c.l.Info("setting TTL for new key", zap.String("key", key))
	}
	return nil
}

func (c *Client) CleanByKey(ctx context.Context, key string) error {
	_, err := c.rdb.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("delete by key error: %w", err)
	}
	c.l.Info("key removed", zap.String("key", key))
	return nil
}
