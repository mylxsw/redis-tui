package api

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mylxsw/redis-tui/config"
	"github.com/mylxsw/redis-tui/core"

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
func NewRedisClient(conf config.Config, outputChan chan core.OutputMessage) RedisClient {
	if conf.Cluster {
		options := &redis.ClusterOptions{
			Addrs:    []string{fmt.Sprintf("%s:%d", conf.Host, conf.Port)},
			Password: conf.Password,
		}

		return redis.NewClusterClient(options)
	}

	options := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		DB:           conf.DB,
		Password:     conf.Password,
		WriteTimeout: 3 * time.Second,
		ReadTimeout:  2 * time.Second,
	}

	client := redis.NewClient(options)
	if conf.Debug {
		client.WrapProcess(func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
			return func(cmd redis.Cmder) error {

				outputChan <- core.OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("redis: <%s>", cmd)}
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

var redisKeys = make([]string, 0)
var redisKeysLastUpdate time.Time
var redisLock sync.RWMutex

func RedisKeys(client RedisClient, pattern string) ([]string, error) {
	keys, err := KeysWithLimit(client, pattern, -1)
	if err != nil {
		return nil, nil
	}

	return keys, nil
}

func RedisAllKeys(client RedisClient, cache bool) ([]string, error) {
	redisLock.RLock()
	if cache && redisKeysLastUpdate.After(time.Now().Add(60*time.Second)) {
		redisLock.RUnlock()
		return redisKeys, nil
	}
	redisLock.RUnlock()

	redisLock.Lock()
	defer redisLock.Unlock()

	keys, err := KeysWithLimit(client, "*", 10)
	if err != nil {
		return nil, err
	}

	redisKeys = keys
	redisKeysLastUpdate = time.Now()

	return keys, nil
}

func KeysWithLimit(client RedisClient, key string, maxScanCount int) (redisKeys []string, err error) {
	var cursor uint64 = 0
	var keys []string

	var scanCount = 0
	for scanCount < maxScanCount || maxScanCount == -1{
		scanCount++

		keys, cursor, err = client.Scan(cursor, key, 100).Result()
		if err != nil {
			return
		}

		redisKeys = append(redisKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	return
}

func RedisServerInfo(conf config.Config, client RedisClient) (string, error) {
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
	if ks, ok := kvpairs[fmt.Sprintf("db%d", conf.DB)]; ok {
		keySpace = ks
	}
	return fmt.Sprintf(" RedisVersion: %s    Memory: %s    Server: %s:%d/%d\n KeySpace: %s", kvpairs["redis_version"], kvpairs["used_memory_human"], conf.Host, conf.Port, conf.DB, keySpace), nil
}
