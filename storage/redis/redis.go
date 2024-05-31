package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
func IncrBy(key string, value int64) *redis.IntCmd {
	return GetRedis().IncrBy(context.Background(), GetCacheKey(key), value)
}

// IncrByFloat
func IncrByFloat(key string, value float64) *redis.FloatCmd {
	return GetRedis().IncrByFloat(context.Background(), GetCacheKey(key), value)
}

// DecrBy
func DecrBy(key string, value int64) *redis.IntCmd {
	return GetRedis().DecrBy(context.Background(), GetCacheKey(key), value)
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
	icr, err := IncrBy(ckey, 1).Result()
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
	return AddQueueByScore(queueKey, i, 0)
}

// AddQueueByScore
func AddQueueByScore(queueKey string, i interface{}, score float64) error {
	byteData, err := json.Marshal(i)
	if err != nil {
		return err
	}

	return GetRedis().ZAdd(context.Background(), GetCacheKey(queueKey), redis.Z{
		Score:  score,
		Member: byteData,
	}).Err()
}

// AddDelayQueue
func AddDelayQueue(queueKey string, i interface{}, delay time.Duration) error {
	return AddQueueByScore(queueKey, i, float64(time.Now().Add(delay).Unix()))
}

// AddPriorityQueue
func AddPriorityQueue(queueKey string, i interface{}, priority int64) error {
	return AddQueueByScore(queueKey, i, float64(priority))
}

// RunQueue
// // 一条一条的处理
// redis.RunQueue("task_queue", 10, "+inf", func(data []byte) (interface{}, error) {
// 	log.Infof("data: %s", string(data))
// 	return nil, nil
// })

// // 批量处理
//
//	redis.RunQueue("task_queue", 10, "+inf", func(data []byte) (interface{}, error) {
//		log.Infof("data: %s", string(data))
//		return nil, nil
//	}, func(results []interface{}) error {
//		// 在这里批量处理上面的返回结果
//		log.Infof("results: %v", results)
//		return nil
//	})
func RunQueue(queueKey string, batchNum int, scoreMax float64, callback func([]byte) (interface{}, error), callbacks ...func([]interface{}) error) {
	queueKey = GetCacheKey(queueKey)
	cRedis := GetRedis()
	ticker := time.NewTicker(1 * time.Second)
	ctx := context.Background()

	for range ticker.C {
		results := []interface{}{}
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < batchNum; i++ {
			// 使用 ZPOPMIN 获取并移动任务
			res, err := cRedis.ZPopMin(ctx, queueKey, 1).Result()
			if err == redis.Nil || len(res) == 0 {
				// 队列为空，继续等待下一次查询
				break
			} else if err != nil {
				log.Errorf("Error popping task from %s: %s", queueKey, err)
				continue
			}

			data := res[0].Member.([]byte)
			score := res[0].Score

			if score > scoreMax {
				// 如果当前元素的分数大于scoreMax，返回队列并跳过
				cRedis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: data})
				continue
			}

			wg.Add(1)
			// 异步执行函数
			go func(data []byte, score float64) {
				defer wg.Done()
				r, err := callback(data)
				if err != nil {
					log.Error(err)
					// 将处理失败的任务重新放回队列
					cRedis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: data})
					return
				}

				if len(callbacks) > 0 {
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				}
			}(data, score)
		}

		// 等待所有异步处理完成
		wg.Wait()

		// 批量操作执行
		if len(results) > 0 {
			for _, cb := range callbacks {
				if err := cb(results); err != nil {
					log.Error(err)
				}
			}
		}
	}
}
