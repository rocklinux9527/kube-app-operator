package db

import (
	"github.com/redis/go-redis/v9"
	"context"
)

func InitRedis(addr string, db int, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return rdb
}

