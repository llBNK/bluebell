package redis

import (
	"bluebell/settings"
	"fmt"

	"github.com/go-redis/redis"
)

const NIl = "redis: nil"

// 声明一个全局的rdb变量
var Rdb *redis.Client

// 初始化连接
func Init(cfg *settings.RedisConfig) (err error) {
	Rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port),
		Password: "",     // no password set
		DB:       cfg.DB, // use default DB
		PoolSize: cfg.PoolSize,
	})

	_, err = Rdb.Ping().Result()
	if err != nil {
		return err
	}
	return
}

func Close() {
	_ = Rdb.Close()
}
