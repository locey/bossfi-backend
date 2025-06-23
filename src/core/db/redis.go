package db

import (
	"bossfi-backend/src/core/config"
	"bossfi-backend/src/core/log"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

var RedisConn *redis.Pool

// InitRedis 初始化Redis
func InitRedis() *redis.Pool {
	log.Logger.Info("Init Redis")
	redisConf := config.Conf.Redis
	// 建立连接池
	RedisConn = &redis.Pool{
		MaxIdle:     10,   // 最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。
		MaxActive:   0,    // 最大的激活连接数，表示同时最多有N个连接   0 表示无穷大
		Wait:        true, // 如果连接数不足则阻塞等待
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisConf.Host, redisConf.Port))
			if err != nil {
				return nil, err
			}
			// 如指定密码，则验证密码
			if strings.TrimSpace(redisConf.Password) != "" {
				_, err = c.Do("auth", redisConf.Password)
				if err != nil {
					panic("redis auth err " + err.Error())
				}
			}
			// 选择db
			_, err = c.Do("select", redisConf.Db)
			if err != nil {
				panic("redis select db err " + err.Error())
			}
			return c, nil
		},
	}
	err := RedisConn.Get().Err()
	if err != nil {
		panic("redis init err " + err.Error())
	}
	return RedisConn
}
