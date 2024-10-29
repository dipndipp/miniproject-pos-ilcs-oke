package utils

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var Ctx = context.Background()

func InitRedis() {
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    _, err := rdb.Ping(Ctx).Result()
    if err != nil {
        log.Fatalf("Tidak dapat terhubung ke Redis: %v", err)
    }
}