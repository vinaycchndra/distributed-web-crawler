package redis_utils

import (
	"context"
	"errors"
	"fmt"

	redis "github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisPassword string
var redisHost string
var ctx context.Context
var RedisConnect *RedisClient

type RedisClient struct {
	redisClient *redis.Client
	ctx         context.Context
}

func InitRedisMem() {
	redisPassword = "password"
	redisHost = "localhost:6379"
	if redisClient == nil {
		ctx = context.Background()
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPassword,
		})
	}

	pong, err := redisClient.Ping(ctx).Result()
	fmt.Println("Testing redis server being up: ", pong, err)

	RedisConnect = &RedisClient{
		redisClient: redisClient,
		ctx:         ctx,
	}
}

func (rC *RedisClient) createBloomFilter(bloomFilterName string) error {
	if bloomFilterName == "" {
		return errors.New("empty bloom filter name")
	}
	_, err := rC.redisClient.Do(rC.ctx, "BF.RESERVE", bloomFilterName, 0.001, 1000000000).Result()

	if err != nil {
		return err
	}
	return nil
}

func (rC *RedisClient) checkBloomFilterExists(bloomFilterName string) (bool, error) {
	if bloomFilterName == "" {
		return false, errors.New("empty bloom filter name")
	}
	_, err := rC.redisClient.Do(rC.ctx, "BF.EXISTS", bloomFilterName, "dummy").Result()

	if err != nil {
		return false, nil
	}
	return true, nil
}

func (rC *RedisClient) LPush(queuName string, data ...any) error {
	return rC.redisClient.LPush(ctx, queuName, data...).Err()
}

func (rC *RedisClient) RPush(queuName string, data ...any) error {
	return rC.redisClient.RPush(ctx, queuName, data...).Err()
}

func (rC *RedisClient) RPop(queuName string) string {
	val, err := rC.redisClient.RPop(ctx, queuName).Result()
	if err != nil {
		return ""
	}
	return val
}

func (rC *RedisClient) LPop(queuName string) any {
	val, err := rC.redisClient.LPop(ctx, queuName).Result()
	if err != nil {
		return ""
	}
	return val
}

func (rC *RedisClient) LLen(queuName string) int64 {
	val, err := rC.redisClient.LLen(ctx, queuName).Result()
	if err != nil {
		return 0
	}
	return val
}
