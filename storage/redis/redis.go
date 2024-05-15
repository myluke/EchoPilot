package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/mylukin/EchoPilot/helper"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Ring

// RedisNil is redis nil
var RedisNil = redis.Nil

// cachePrefix
var cachePrefix = "%s"

// 创建一个 redis 连接
func GetRedis() *redis.Ring {
	if redisClient != nil {
		return redisClient
	}
	redisDB, _ := strconv.Atoi(helper.Config("REDIS_DB"))
	redisServers := helper.Config("REDIS_SERVERS")
	redisAddrs := map[string]string{}
	for _, v := range strings.Split(redisServers, ",") {
		k := helper.MD5(v)
		redisAddrs[k] = strings.TrimSpace(v)
	}
	redisClient = redis.NewRing(&redis.RingOptions{
		Addrs:    redisAddrs,
		Password: helper.Config("REDIS_PASSWORD"), // no password set
		DB:       redisDB,                         // use default DB
	})
	return redisClient
}

// Prefix is set prefix
func Prefix(key string) {
	cachePrefix = fmt.Sprintf("%s:%%s", key)
}

// GetPrefix is get prefix
func GetPrefix() string {
	return cachePrefix
}

// GetCacheKey is get cache key
func GetCacheKey(key string) string {
	return fmt.Sprintf(cachePrefix, key)
}

// Expire
func Expire(key string, expiration time.Duration) *redis.BoolCmd {
	return GetRedis().Expire(context.Background(), GetCacheKey(key), expiration)
}

// Set
func Set(key string, value any, expiration time.Duration) *redis.StatusCmd {
	return GetRedis().Set(context.Background(), GetCacheKey(key), value, expiration)
}

// JsonSet
func JsonSet(key string, value any, expiration time.Duration) *redis.StatusCmd {
	ctx := context.Background()
	redisObj := GetRedis()
	cacheKey := GetCacheKey(key)
	resp := redisObj.JSONSet(ctx, cacheKey, "$", value)
	if resp.Err() == nil {
		redisObj.Expire(ctx, cacheKey, expiration)
	}
	return resp
}

// IncrBy
func IncrBy(key string, value int64) (int64, error) {
	return GetRedis().IncrBy(context.Background(), GetCacheKey(key), value).Result()
}

// IncrByFloat
func IncrByFloat(key string, value float64) (float64, error) {
	return GetRedis().IncrByFloat(context.Background(), GetCacheKey(key), value).Result()
}

// DecrBy
func DecrBy(key string, value int64) (int64, error) {
	return GetRedis().DecrBy(context.Background(), GetCacheKey(key), value).Result()
}

// Get
func Get(key string) *redis.StringCmd {
	return GetRedis().Get(context.Background(), GetCacheKey(key))
}

// JsonSet
func JsonGet(key string, paths ...string) *redis.JSONCmd {
	return GetRedis().JSONGet(context.Background(), GetCacheKey(key), paths...)
}

// Has
func Has(key string) bool {
	_, err := Get(key).Result()
	return err != redis.Nil
}

// Del
func Del(keys ...string) *redis.IntCmd {
	newKeys := []string{}
	for _, k := range keys {
		newKeys = append(newKeys, GetCacheKey(k))
	}
	return GetRedis().Del(context.Background(), newKeys...)
}

// TTL
func TTL(key string) *redis.DurationCmd {
	return GetRedis().TTL(context.Background(), GetCacheKey(key))
}

// AnyDo
func AnyDo(name string, expiration time.Duration) int {
	ckey := fmt.Sprintf("EveryAnyDo:%s:%s", name, expiration.String())
	icr, err := IncrBy(ckey, 1)
	if err != nil {
		return 0
	}
	if icr == 1 {
		Expire(ckey, expiration)
	}
	return int(icr)
}

// HourDo 每小时可执行次数
func HourDo(name string) int {
	return AnyDo(name, 1*time.Hour)
}

// DayDo 每天可执行次数
func DayDo(name string) int {
	return AnyDo(name, 24*time.Hour)
}

// MonthDo 每月可执行次数
func MonthDo(name string) int {
	return AnyDo(name, 30*24*time.Hour)
}

// AddQueue
func AddQueue(queueKey string, i interface{}) error {
	return AddDelayQueue(queueKey, i, 0)
}

// AddDelayQueue
func AddDelayQueue(queueKey string, i interface{}, delay time.Duration) error {
	byteData, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return GetRedis().ZAdd(context.Background(), GetCacheKey(queueKey), redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: byteData,
	}).Err()
}

// RunQueue
// // 一条一条的处理
// redis.RunQueue("task_queue", 10, func(data []byte) (interface{}, error) {
// 	log.Infof("data: %s", string(data))
// 	return nil, nil
// })

// // 批量处理
//
//	redis.RunQueue("task_queue", 10, func(data []byte) (interface{}, error) {
//		log.Infof("data: %s", string(data))
//		return nil, nil
//	}, func(results []interface{}) error {
//		// 在这里批量处理上面的返回结果
//		log.Infof("results: %v", results)
//		return nil
//	})
func RunQueue(queueKey string, batchNum int, callback func([]byte) (interface{}, error), callbacks ...func([]interface{}) error) {
	queueKey = GetCacheKey(queueKey)
	cRedis := GetRedis()
	ticker := time.NewTicker(1 * time.Second)
	ctx := context.Background()

	for range ticker.C {
		results := []interface{}{}
		res, err := cRedis.ZRangeByScoreWithScores(ctx, queueKey, &redis.ZRangeBy{
			Min:    "-inf",
			Max:    fmt.Sprintf("%d", time.Now().Unix()),
			Offset: 0,
			Count:  int64(batchNum),
		}).Result()

		// 处理失败，跳过
		if err != nil {
			// 队列为空，退出
			if err == RedisNil {
				// 队列为空，继续等待下一次查询
				continue
			}
			log.Error(err)
			continue
		}

		if len(res) == 0 {
			// 没有到期任务，继续等待下一次查询
			continue
		}

		// log.Infof("res: %v", res)

		members := []interface{}{}
		for _, rz := range res {
			// 到了执行时间
			data := rz.Member.(string)

			// 执行函数
			r, err := callback([]byte(data))
			if err != nil {
				log.Error(err)
				continue
			}

			members = append(members, rz.Member)

			if len(callbacks) > 0 {
				results = append(results, r)
			}
		}

		// 移除已成功处理的数据
		if len(members) > 0 {
			err := cRedis.ZRem(ctx, queueKey, members...)
			if err != nil {
				log.Error("Error removing processed members: ", err)
			} else {
				log.Infof("Removed processed members: %v", members)
			}
		}

		// 批量操作执行
		for _, cb := range callbacks {
			if err := cb(results); err != nil {
				log.Error(err)
			}
		}
	}
}
