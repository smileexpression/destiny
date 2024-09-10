package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"smile.expression/destiny/logger"
)

type Client struct {
	redisClient *redis.Client
	options     *Options
}

type Options struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Duration int    `json:"duration"`
}

func NewClient(options *Options) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     options.Addr,
		Password: options.Password,
		DB:       options.DB,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(err)
	}

	return &Client{
		redisClient: client,
		options:     options,
	}
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	var (
		log = logger.SmileLog.WithContext(ctx)
	)

	data, err := c.redisClient.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		log.WithError(err).Infof("get cache miss: %s", key)
		return nil, err
	}

	if err != nil {
		log.WithError(err).Errorf("get cache fail: %s", key)
		return nil, err
	}

	log.Infof("get cache success: %s", key)
	return data, nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte, expiration int) error {
	var (
		log = logger.SmileLog.WithContext(ctx)
	)

	if expiration < 0 {
		expiration = c.options.Duration
	}

	if err := c.redisClient.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err(); err != nil {
		return err
	}
	log.Infof("set cache success: %s", key)
	return nil
}
