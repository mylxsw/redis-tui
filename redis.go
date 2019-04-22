package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"

	"github.com/go-redis/redis"
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
	Do(args ...interface{}) *redis.Cmd
	Info(section ...string) *redis.StringCmd
}

// NewRedisClient create a new redis client which wraps single or cluster client
func NewRedisClient(config Config, outputChan chan OutputMessage) RedisClient {
	if config.Cluster {
		options := &redis.ClusterOptions{
			Addrs:    []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
			Password: config.Password,
		}

		return redis.NewClusterClient(options)
	}

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:       config.DB,
		Password: config.Password,
	}

	client := redis.NewClient(options)
	if config.Debug {
		client.WrapProcess(func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
			return func(cmd redis.Cmder) error {

				outputChan <- OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("redis: <%s>", cmd)}
				err := oldProcess(cmd)

				return err
			}
		})
	}

	return client
}

func RedisExecute(client RedisClient, command string) (interface{}, error) {
	stringArgs := strings.Split(command, " ")
	var args = make([]interface{}, len(stringArgs))
	for i, s := range stringArgs {
		args[i] = s
	}

	return client.Do(args...).Result()
}

func RedisServerInfo(config Config, client RedisClient) (string, error) {
	res, err := client.Info().Result()
	if err != nil {
		return "", err
	}

	var kvpairs = make(map[string]string)
	for _, kv := range strings.Split(res, "\n") {
		if strings.HasPrefix(kv, "#") || kv == "" {
			continue
		}

		pair := strings.SplitN(kv, ":", 2)
		if len(pair) != 2 {
			continue
		}

		kvpairs[pair[0]] = pair[1]
	}

	keySpace := "-"
	if ks, ok := kvpairs[fmt.Sprintf("db%d", config.DB)]; ok {
		keySpace = ks
	}
	return fmt.Sprintf(" RedisVersion: %s    Memory: %s    Server: %s:%d/%d\n KeySpace: %s", kvpairs["redis_version"], kvpairs["used_memory_human"], config.Host, config.Port, config.DB, keySpace), nil
}
