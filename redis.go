package main

import (
	"fmt"
	"gopkg.in/redis.v5"
)

type RedisClient interface {
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
	Type(key string) *redis.StatusCmd
	TTL(key string) *redis.DurationCmd
	Get(key string) *redis.StringCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	SMembers(key string) *redis.StringSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	HKeys(key string) *redis.StringSliceCmd
	HGet(key, field string) *redis.StringCmd
}

// NewRedisClient create a new redis client which wraps single or cluster client
func NewRedisClient(config Config) RedisClient {
	if config.Cluster {
		options := &redis.ClusterOptions{
			Addrs:    []string{fmt.Sprintf("%s:%d", config.Host, config.Port),},
			Password: config.Password,
		}

		return redis.NewClusterClient(options)
	}

	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:   config.DB,
	}

	client := redis.NewClient(options)

	return client
}
