package redis_utils

import (
	"fmt"

	redisbloom "github.com/RedisBloom/redisbloom-go"
)

var bloomFilterName = "crawlerWithBloomFilter"
var redisBloomClient *redisbloom.Client
var BloomFilterClient *BloomFilter

type BloomFilter struct {
	bf *redisbloom.Client
}

func InitRedisBloom() {
	redisPassword = "password"
	redisHost = "localhost:6379"
	exists, err := RedisConnect.checkBloomFilterExists(bloomFilterName)

	if err != nil {
		panic("error while checking the existence of bloom filter")
	} else {
		fmt.Println("Bloom Filter created")
	}

	if !exists {
		err = RedisConnect.createBloomFilter(bloomFilterName)
		if err != nil {
			panic("error while creating bloom filter")
		}
	} else {
		fmt.Println("Bloom filter exists")
	}

	if redisBloomClient == nil {
		redisBloomClient = redisbloom.NewClient(redisHost, "crawlBloom", &redisPassword)
	}

	BloomFilterClient = &BloomFilter{
		bf: redisBloomClient,
	}
}

// Returns the bool if the key is created or not..., key will be created if it does not exists.
func (bF *BloomFilter) Add(key string) (bool, error) {
	ok, err := bF.bf.Add(bloomFilterName, key)

	if err != nil {
		fmt.Printf("failed to add item : %s to the bloom filter key : %s   error: %s \n", key, bloomFilterName, err.Error())
		return false, err
	}
	return ok, nil
}
