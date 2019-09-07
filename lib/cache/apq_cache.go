/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package cache

import (
	"context"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"time"
)

//Cache Automatic persisted queries
type Cache struct {
	client redis.UniversalClient
	ttl    time.Duration
}

const apqPrefix = "apq:"

//NewAPQCache create new automatic persisted queries cache for caching graphql queries
func NewAPQCache(redisClient *redis.Client, ttl time.Duration) (*Cache, error) {

	err := redisClient.Ping().Err()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Cache{client: redisClient, ttl: ttl}, nil
}

func (c *Cache) Add(ctx context.Context, hash string, query string) {
	c.client.Set(apqPrefix+hash, query, c.ttl)
}

func (c *Cache) Get(ctx context.Context, hash string) (string, bool) {
	s, err := c.client.Get(apqPrefix + hash).Result()
	if err != nil {
		return "", false
	}
	return s, true
}
