package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strconv"
)

// Config 保存运行时配置
type Config struct {
	// MySQL DSN，例如： root:rootpassword@tcp(127.0.0.1:3306)/approval_db?charset=utf8mb4&parseTime=True&loc=Local
	MySQLDSN string
	// Redis 地址，例如：127.0.0.1:6379
	RedisAddr string
	// Redis DB 索引，默认为 0
	RedisDB int
        RedisPassword string // Redis 密码
}

// Load 从环境变量读取配置（有默认值）
func Load() *Config {
	cfg := &Config{
                 //loc=Local 
                // Asia%2FShanghai
		MySQLDSN:   getEnv("MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/k8s?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"),
		RedisAddr:  getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisDB:    atoiOrDefault(getEnv("REDIS_DB", "0"), 0),
		RedisPassword: getEnv("REDIS_PASSWORD", "I51JMOUEuY"),
	}
	return cfg
}

// getEnv 读取 env 或返回默认值
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// atoiOrDefault 尝试转换字符串为 int，失败返回默认值
func atoiOrDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// InitMySQL 初始化 MySQL 连接

func InitMySQL(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}

// InitRedis 初始化 Redis 客户端
func InitRedis(cfg *Config) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       cfg.RedisDB,
    })
}

// PingRedis 测试 Redis 连接
func PingRedis(rdb *redis.Client) error {
    return rdb.Ping(context.Background()).Err()
}
