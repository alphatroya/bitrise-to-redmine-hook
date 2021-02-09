package main

import (
	"time"

	"github.com/go-redis/redis"
)

type Storage interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
}
