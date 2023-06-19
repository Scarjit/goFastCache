package cache

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/redis/go-redis/v9"
	"goFastCache/pkg/hash"
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

func (c *Cache) GetList(domain, user, repo string) (string, error) {
	key := hash.GetHash(domain, user, repo)
	val, err := c.redis.Get(context.Background(), hex.EncodeToString(key[:])).Result()
	return val, err
}

func (c *Cache) SetList(domain, user, repo, list string) error {
	key := hash.GetHash(domain, user, repo)
	return c.redis.SetArgs(context.Background(), hex.EncodeToString(key[:]), list, redis.SetArgs{
		ExpireAt: time.Now().Add(time.Minute * 1),
	}).Err()
}

func (c *Cache) GetSumObj(domain, trail string) (string, error) {
	key := hash.GetMinioSumPath(domain, trail)
	val, err := c.redis.Get(context.Background(), key).Result()
	return val, err
}

func (c *Cache) SetSumObj(domain, trail string, sum []byte) error {
	key := hash.GetMinioSumPath(domain, trail)
	return c.redis.SetArgs(context.Background(), key, sum, redis.SetArgs{
		ExpireAt: time.Now().Add(time.Minute * 10),
	}).Err()
}
