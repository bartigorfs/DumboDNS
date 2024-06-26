package database

import "github.com/redis/go-redis/v9"

var LocalDNSSub *redis.PubSub

func RedisCacheClient() *redis.Client {
	url := "redis://localhost:6379/1?protocol=3"
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opts)
}
