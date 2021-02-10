package main

import (
	"time"

	"github.com/go-redis/redis"
)

// Storage represents interface for storing data in external vault
type Storage interface {
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
}
