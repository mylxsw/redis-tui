package main

import (
	"fmt"
	"gopkg.in/redis.v5"
	"strings"
)

type RedisClient interface {
	Keys(pattern string) *redis.StringSliceCmd
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
	Type(key string) *redis.StatusCmd
	TTL(key string) *redis.DurationCmd
	Get(key string) *redis.StringCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	SMembers(key string) *redis.StringSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	HKeys(key string) *redis.StringSliceCmd
	HGet(key, field string) *redis.StringCmd
	Process(cmd redis.Cmder) error
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

func RedisExecute(client RedisClient, command string) (interface{}, error) {
	stringArgs := strings.Split(command, " ")
	var args = make([]interface{}, len(stringArgs))
	for i, s := range stringArgs {
		args[i] = s
	}

	cmd := redis.NewCmd(args...)
	err := client.Process(cmd)
	if err != nil {
		return nil, err
	}

	return cmd.Result()
}
