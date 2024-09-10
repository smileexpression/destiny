package cache

import (
	"context"
	"errors"

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
		log.WithError(err).Info("cache miss")
		return nil, err
	}

	if err != nil {
		log.WithError(err).Error("get cache fail")
		return nil, err
	}

	log.Info("get cache success")
	return data, nil
}

func (c *Client) Set(ctx context.Context, key string, value []byte) error {
	var (
		log = logger.SmileLog.WithContext(ctx)
	)

	if err := c.redisClient.Set(ctx, key, value, 0).Err(); err != nil {
		return err
	}
	log.Info("set cache success")
	return nil
}
