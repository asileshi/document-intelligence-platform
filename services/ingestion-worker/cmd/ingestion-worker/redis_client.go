package main

import "github.com/redis/go-redis/v9"

func newRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}
