package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"os"
	"strings"
	"time"
)

type Cache struct {
	redis *redis.Client
}

func NewCache() (*Cache, error) {
	redisAddress, found := os.LookupEnv("REDIS_ADDRESS")
	if !found {
		return nil, errors.New("REDIS_ADDRESS not found")
	}
	redisAddress = strings.Trim(redisAddress, "\n\r")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	ping := rdb.Ping(context.Background())
	if ping.Err() != nil {
		return nil, ping.Err()
	}

	return &Cache{
		redis: rdb,
	}, nil
}

func (c *Cache) Set(key string, value []byte, expiresIn time.Duration) (err error) {
	status := c.redis.SetArgs(context.Background(), key, value, redis.SetArgs{
		ExpireAt: time.Now().Add(expiresIn),
	})
	if status != nil {
		return status.Err()
	}
	return nil
}

func (c *Cache) Get(key string) ([]byte, bool, error) {
	status := c.redis.Get(context.Background(), key)
	if status.Err() != nil {
		return nil, false, status.Err()
	}
	bytes, err := status.Bytes()
	if err != nil {
		return nil, false, err
	}
	return bytes, true, nil
}
