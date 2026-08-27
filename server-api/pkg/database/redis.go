package database

import (
	"context"
	"megin/pkg/logger"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// 基于go-redis v8封装
type RedisClient struct {
	rdb *redis.Client
	ctx context.Context
}

var redisClient = new(RedisClient)

func RedisConnect(addr, password string) *RedisClient {
	if len(addr) <= 0 {
		addr = "127.0.0.1:6379"
	}

	redisClient.rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	redisClient.ctx = context.Background()

	err := Set("test", 1, time.Minute)
	if err != nil {
		logger.Fatal("Redis Error", zap.Error(err))
	}
	return redisClient
}

func (this RedisClient) GetDB() *redis.Client {
	return this.rdb
}

// 不过期
func SetForever(key string, value any) error {
	return redisClient.rdb.Set(redisClient.ctx, key, value, 0).Err()
}

// 设置过期时间,为了防把过期时间误传,如果不到1秒的,会当成秒处理
func Set(key string, value any, expiration time.Duration) error {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	return redisClient.rdb.Set(redisClient.ctx, key, value, expiration).Err()
}

func LPush(key string, value any, expiration time.Duration) error {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	err := redisClient.rdb.LPush(redisClient.ctx, key, value).Err()
	if err != nil {
		return err
	}
	err = redisClient.rdb.Expire(redisClient.ctx, key, expiration).Err()
	return err
}

func LPushForever(key string, value any) error {
	err := redisClient.rdb.LPush(redisClient.ctx, key, value).Err()
	return err
}

func RPop(key string) (string, error) {
	cmd := redisClient.rdb.RPop(redisClient.ctx, key)
	return cmd.Result()
}

func RangeAll(key string) ([]string, error) {
	cmd := redisClient.rdb.LRange(redisClient.ctx, key, 0, -1)
	var list []string
	if cmd.Err() != nil {
		return list, cmd.Err()
	}
	list = cmd.Val()
	return list, nil
}

func Lock(key string, value any, expiration time.Duration) bool {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	cmd := redisClient.rdb.SetNX(redisClient.ctx, key, value, expiration)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

func LockForever(key string, value int) bool {
	cmd := redisClient.rdb.SetNX(redisClient.ctx, key, value, 0)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

// GET
func Get(key string) *redis.StringCmd {
	return redisClient.rdb.Get(redisClient.ctx, key)
}

//key:string,value:struct
func HMSet(key string, values any, expiration time.Duration) *redis.BoolCmd {
	if expiration < time.Second {
		expiration = expiration * time.Second
	}
	return redisClient.rdb.HMSet(redisClient.ctx, key, values)
}

func HMGet(key string) *redis.StringStringMapCmd {
	return redisClient.rdb.HGetAll(redisClient.ctx, key)
}

//key:string,value:struct
func HSet(key string, field string, values any) *redis.IntCmd {
	return redisClient.rdb.HSet(redisClient.ctx, key, field, values)
}

func HDel(key string, fields ...string) *redis.IntCmd {
	return redisClient.rdb.HDel(redisClient.ctx, key, fields...)
}

func HGetAll(key string) *redis.StringStringMapCmd {
	return redisClient.rdb.HGetAll(redisClient.ctx, key)
}

func HGet(key, field string) *redis.StringCmd {
	return redisClient.rdb.HGet(redisClient.ctx, key, field)
}

func IncrFloat(key string, val float64) *redis.FloatCmd {
	return redisClient.rdb.IncrByFloat(redisClient.ctx, key, val)
}

func Incr(key string) *redis.IntCmd {
	return redisClient.rdb.Incr(redisClient.ctx, key)
}

func Delete(key string) *redis.IntCmd {
	return redisClient.rdb.Del(redisClient.ctx, key)
}
