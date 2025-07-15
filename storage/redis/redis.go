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

// GetAnyDoKey
func GetAnyDoKey(name string, expiration time.Duration) string {
	return fmt.Sprintf("EveryAnyDo:%s:%s", name, expiration.String())
}

// AnyDo
func AnyDo(name string, expiration time.Duration) int {
	ckey := GetAnyDoKey(name, expiration)
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

// RunQueue processes tasks from a Redis sorted set queue.
//
// Example usage:
//
// 1. Process tasks one by one:
//
//	redis.RunQueue("task_queue", 10, "inf", func(data []byte) (interface{}, error) {
//		log.Infof("Processing task: %s", string(data))
//		// Process the task here
//		return nil, nil
//	})
//
// 2. Process tasks in batches:
//
//	redis.RunQueue("task_queue", 10, "inf", func(data []byte) (interface{}, error) {
//		// Process individual task
//		return processTask(data), nil
//	}, func(results []interface{}) error {
//		// Batch process the results
//		return batchProcessResults(results)
//	})
//
// 3. Process time-based tasks:
//
//	redis.RunQueue("delayed_task_queue", 10, "time", func(data []byte) (interface{}, error) {
//		log.Infof("Processing delayed task: %s", string(data))
//		// Process the delayed task here
//		return nil, nil
//	})
//
// Parameters:
//   - queueKey: The key of the Redis sorted set queue
//   - batchNum: The maximum number of tasks to process in each iteration
//   - qType: Use "inf" for regular queue, "time" for time-based queue
//   - callback: Function to process each task
//   - callbacks: Optional function(s) for batch processing results
func RunQueue(queueKey string, batchNum int, qType string, callback func([]byte) (interface{}, error), callbacks ...func([]interface{}) error) {
	queueKey = GetCacheKey(queueKey)
	cRedis := GetRedis()
	ticker := time.NewTicker(1 * time.Second)
	ctx := context.Background()

	isTimeCheck := qType == "time"

	for range ticker.C {
		results := []interface{}{}
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 0; i < batchNum; i++ {
			// Use ZPOPMIN to get and move tasks
			res, err := cRedis.ZPopMin(ctx, queueKey, 1).Result()
			if err == redis.Nil || len(res) == 0 {
				// Queue is empty, continue waiting for the next query
				break
			} else if err != nil {
				log.Errorf("Error popping task from %s: %s", queueKey, err)
				continue
			}

			data := res[0].Member.(string)
			score := res[0].Score

			if isTimeCheck {
				currentTime := float64(time.Now().Unix())
				if score > currentTime {
					// If the current element's time is later than the current time, return to the queue and skip
					cRedis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: res[0].Member})
					continue
				}
			}

			wg.Add(1)
			// Execute function asynchronously
			go func(data string, score float64) {
				defer wg.Done()
				r, err := callback([]byte(data))
				if err != nil {
					log.Error(err)
					// Put the failed task back into the queue
					cRedis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: res[0].Member})
					return
				}

				if len(callbacks) > 0 {
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				}
			}(data, score)
		}

		// Wait for all asynchronous processing to complete
		wg.Wait()

		// Execute batch operations
		if len(results) > 0 {
			for _, cb := range callbacks {
				if err := cb(results); err != nil {
					log.Error(err)
				}
			}
		}
	}
}

// ============================================================================
// 批量操作 - Batch Operations
// ============================================================================

// MSet sets multiple key-value pairs
func MSet(pairs ...interface{}) *redis.StatusCmd {
	ctx := context.Background()
	// Convert keys to have cache prefix
	newPairs := make([]interface{}, len(pairs))
	for i := 0; i < len(pairs); i += 2 {
		if i+1 < len(pairs) {
			newPairs[i] = GetCacheKey(pairs[i].(string))
			newPairs[i+1] = pairs[i+1]
		}
	}
	return GetRedis().MSet(ctx, newPairs...)
}

// MGet gets multiple values by keys
func MGet(keys ...string) *redis.SliceCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().MGet(ctx, newKeys...)
}

// Exists checks if keys exist
func Exists(keys ...string) *redis.IntCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().Exists(ctx, newKeys...)
}

// Keys returns all keys matching pattern
func Keys(pattern string) *redis.StringSliceCmd {
	return GetRedis().Keys(context.Background(), GetCacheKey(pattern))
}

// Scan iterates over keys
func Scan(cursor uint64, match string, count int64) *redis.ScanCmd {
	return GetRedis().Scan(context.Background(), cursor, GetCacheKey(match), count)
}

// ============================================================================
// 字符串操作扩展 - String Operations Extended
// ============================================================================

// SetNX sets key to value if key does not exist
func SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return GetRedis().SetNX(context.Background(), GetCacheKey(key), value, expiration)
}

// SetEX sets key to value with expiration
func SetEX(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return GetRedis().SetEx(context.Background(), GetCacheKey(key), value, expiration)
}

// GetSet atomically sets key to value and returns the old value
func GetSet(key string, value interface{}) *redis.StringCmd {
	return GetRedis().GetSet(context.Background(), GetCacheKey(key), value)
}

// Append appends value to key
func Append(key string, value string) *redis.IntCmd {
	return GetRedis().Append(context.Background(), GetCacheKey(key), value)
}

// StrLen returns the length of string stored at key
func StrLen(key string) *redis.IntCmd {
	return GetRedis().StrLen(context.Background(), GetCacheKey(key))
}

// Incr increments key by 1
func Incr(key string) *redis.IntCmd {
	return GetRedis().Incr(context.Background(), GetCacheKey(key))
}

// Decr decrements key by 1
func Decr(key string) *redis.IntCmd {
	return GetRedis().Decr(context.Background(), GetCacheKey(key))
}

// ============================================================================
// 哈希操作 - Hash Operations
// ============================================================================

// HSet sets field in hash stored at key to value
func HSet(key string, values ...interface{}) *redis.IntCmd {
	return GetRedis().HSet(context.Background(), GetCacheKey(key), values...)
}

// HGet returns the value of field in hash stored at key
func HGet(key string, field string) *redis.StringCmd {
	return GetRedis().HGet(context.Background(), GetCacheKey(key), field)
}

// HMSet sets multiple fields in hash stored at key
func HMSet(key string, values ...interface{}) *redis.BoolCmd {
	return GetRedis().HMSet(context.Background(), GetCacheKey(key), values...)
}

// HMGet returns values of multiple fields in hash stored at key
func HMGet(key string, fields ...string) *redis.SliceCmd {
	return GetRedis().HMGet(context.Background(), GetCacheKey(key), fields...)
}

// HGetAll returns all fields and values in hash stored at key
func HGetAll(key string) *redis.MapStringStringCmd {
	return GetRedis().HGetAll(context.Background(), GetCacheKey(key))
}

// HDel deletes one or more fields from hash stored at key
func HDel(key string, fields ...string) *redis.IntCmd {
	return GetRedis().HDel(context.Background(), GetCacheKey(key), fields...)
}

// HExists checks if field exists in hash stored at key
func HExists(key string, field string) *redis.BoolCmd {
	return GetRedis().HExists(context.Background(), GetCacheKey(key), field)
}

// HLen returns the number of fields in hash stored at key
func HLen(key string) *redis.IntCmd {
	return GetRedis().HLen(context.Background(), GetCacheKey(key))
}

// HKeys returns all fields in hash stored at key
func HKeys(key string) *redis.StringSliceCmd {
	return GetRedis().HKeys(context.Background(), GetCacheKey(key))
}

// HVals returns all values in hash stored at key
func HVals(key string) *redis.StringSliceCmd {
	return GetRedis().HVals(context.Background(), GetCacheKey(key))
}

// HIncrBy increments the integer value of field in hash stored at key by increment
func HIncrBy(key string, field string, incr int64) *redis.IntCmd {
	return GetRedis().HIncrBy(context.Background(), GetCacheKey(key), field, incr)
}

// HIncrByFloat increments the float value of field in hash stored at key by increment
func HIncrByFloat(key string, field string, incr float64) *redis.FloatCmd {
	return GetRedis().HIncrByFloat(context.Background(), GetCacheKey(key), field, incr)
}

// ============================================================================
// 列表操作 - List Operations
// ============================================================================

// LPush prepends one or more values to list stored at key
func LPush(key string, values ...interface{}) *redis.IntCmd {
	return GetRedis().LPush(context.Background(), GetCacheKey(key), values...)
}

// RPush appends one or more values to list stored at key
func RPush(key string, values ...interface{}) *redis.IntCmd {
	return GetRedis().RPush(context.Background(), GetCacheKey(key), values...)
}

// LPop removes and returns the first element of list stored at key
func LPop(key string) *redis.StringCmd {
	return GetRedis().LPop(context.Background(), GetCacheKey(key))
}

// RPop removes and returns the last element of list stored at key
func RPop(key string) *redis.StringCmd {
	return GetRedis().RPop(context.Background(), GetCacheKey(key))
}

// LLen returns the length of list stored at key
func LLen(key string) *redis.IntCmd {
	return GetRedis().LLen(context.Background(), GetCacheKey(key))
}

// LRange returns elements from list stored at key
func LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return GetRedis().LRange(context.Background(), GetCacheKey(key), start, stop)
}

// LIndex returns the element at index in list stored at key
func LIndex(key string, index int64) *redis.StringCmd {
	return GetRedis().LIndex(context.Background(), GetCacheKey(key), index)
}

// LSet sets the element at index in list stored at key to value
func LSet(key string, index int64, value interface{}) *redis.StatusCmd {
	return GetRedis().LSet(context.Background(), GetCacheKey(key), index, value)
}

// LTrim trims list stored at key to specified range
func LTrim(key string, start, stop int64) *redis.StatusCmd {
	return GetRedis().LTrim(context.Background(), GetCacheKey(key), start, stop)
}

// ============================================================================
// 集合操作 - Set Operations
// ============================================================================

// SAdd adds one or more members to set stored at key
func SAdd(key string, members ...interface{}) *redis.IntCmd {
	return GetRedis().SAdd(context.Background(), GetCacheKey(key), members...)
}

// SRem removes one or more members from set stored at key
func SRem(key string, members ...interface{}) *redis.IntCmd {
	return GetRedis().SRem(context.Background(), GetCacheKey(key), members...)
}

// SMembers returns all members of set stored at key
func SMembers(key string) *redis.StringSliceCmd {
	return GetRedis().SMembers(context.Background(), GetCacheKey(key))
}

// SIsMember checks if member is a member of set stored at key
func SIsMember(key string, member interface{}) *redis.BoolCmd {
	return GetRedis().SIsMember(context.Background(), GetCacheKey(key), member)
}

// SCard returns the cardinality of set stored at key
func SCard(key string) *redis.IntCmd {
	return GetRedis().SCard(context.Background(), GetCacheKey(key))
}

// SPop removes and returns one or more random members from set stored at key
func SPop(key string) *redis.StringCmd {
	return GetRedis().SPop(context.Background(), GetCacheKey(key))
}

// SPopN removes and returns count random members from set stored at key
func SPopN(key string, count int64) *redis.StringSliceCmd {
	return GetRedis().SPopN(context.Background(), GetCacheKey(key), count)
}

// SRandMember returns one or more random members from set stored at key
func SRandMember(key string) *redis.StringCmd {
	return GetRedis().SRandMember(context.Background(), GetCacheKey(key))
}

// SRandMemberN returns count random members from set stored at key
func SRandMemberN(key string, count int64) *redis.StringSliceCmd {
	return GetRedis().SRandMemberN(context.Background(), GetCacheKey(key), count)
}

// SUnion returns the union of sets stored at keys
func SUnion(keys ...string) *redis.StringSliceCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().SUnion(ctx, newKeys...)
}

// SInter returns the intersection of sets stored at keys
func SInter(keys ...string) *redis.StringSliceCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().SInter(ctx, newKeys...)
}

// SDiff returns the difference of sets stored at keys
func SDiff(keys ...string) *redis.StringSliceCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().SDiff(ctx, newKeys...)
}

// ============================================================================
// 有序集合操作 - Sorted Set Operations
// ============================================================================

// ZAdd adds one or more members to sorted set stored at key
func ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	return GetRedis().ZAdd(context.Background(), GetCacheKey(key), members...)
}

// ZRem removes one or more members from sorted set stored at key
func ZRem(key string, members ...interface{}) *redis.IntCmd {
	return GetRedis().ZRem(context.Background(), GetCacheKey(key), members...)
}

// ZScore returns the score of member in sorted set stored at key
func ZScore(key string, member string) *redis.FloatCmd {
	return GetRedis().ZScore(context.Background(), GetCacheKey(key), member)
}

// ZRank returns the rank of member in sorted set stored at key
func ZRank(key string, member string) *redis.IntCmd {
	return GetRedis().ZRank(context.Background(), GetCacheKey(key), member)
}

// ZRevRank returns the reverse rank of member in sorted set stored at key
func ZRevRank(key string, member string) *redis.IntCmd {
	return GetRedis().ZRevRank(context.Background(), GetCacheKey(key), member)
}

// ZRange returns elements from sorted set stored at key
func ZRange(key string, start, stop int64) *redis.StringSliceCmd {
	return GetRedis().ZRange(context.Background(), GetCacheKey(key), start, stop)
}

// ZRevRange returns elements from sorted set stored at key in reverse order
func ZRevRange(key string, start, stop int64) *redis.StringSliceCmd {
	return GetRedis().ZRevRange(context.Background(), GetCacheKey(key), start, stop)
}

// ZRangeWithScores returns elements with scores from sorted set stored at key
func ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return GetRedis().ZRangeWithScores(context.Background(), GetCacheKey(key), start, stop)
}

// ZRevRangeWithScores returns elements with scores from sorted set stored at key in reverse order
func ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return GetRedis().ZRevRangeWithScores(context.Background(), GetCacheKey(key), start, stop)
}

// ZCard returns the cardinality of sorted set stored at key
func ZCard(key string) *redis.IntCmd {
	return GetRedis().ZCard(context.Background(), GetCacheKey(key))
}

// ZCount returns the number of members in sorted set stored at key with scores between min and max
func ZCount(key string, min, max string) *redis.IntCmd {
	return GetRedis().ZCount(context.Background(), GetCacheKey(key), min, max)
}

// ZIncrBy increments the score of member in sorted set stored at key by increment
func ZIncrBy(key string, increment float64, member string) *redis.FloatCmd {
	return GetRedis().ZIncrBy(context.Background(), GetCacheKey(key), increment, member)
}

// ============================================================================
// 位图操作 - Bitmap Operations
// ============================================================================

// SetBit sets or clears the bit at offset in the string value stored at key
func SetBit(key string, offset int64, value int) *redis.IntCmd {
	return GetRedis().SetBit(context.Background(), GetCacheKey(key), offset, value)
}

// GetBit returns the bit value at offset in the string value stored at key
func GetBit(key string, offset int64) *redis.IntCmd {
	return GetRedis().GetBit(context.Background(), GetCacheKey(key), offset)
}

// BitCount counts the number of set bits in a string
func BitCount(key string, bitCount *redis.BitCount) *redis.IntCmd {
	return GetRedis().BitCount(context.Background(), GetCacheKey(key), bitCount)
}

// BitOpAnd performs bitwise AND operation between strings
func BitOpAnd(destKey string, keys ...string) *redis.IntCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().BitOpAnd(ctx, GetCacheKey(destKey), newKeys...)
}

// BitOpOr performs bitwise OR operation between strings
func BitOpOr(destKey string, keys ...string) *redis.IntCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().BitOpOr(ctx, GetCacheKey(destKey), newKeys...)
}

// BitOpXor performs bitwise XOR operation between strings
func BitOpXor(destKey string, keys ...string) *redis.IntCmd {
	ctx := context.Background()
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = GetCacheKey(key)
	}
	return GetRedis().BitOpXor(ctx, GetCacheKey(destKey), newKeys...)
}

// BitOpNot performs bitwise NOT operation on string
func BitOpNot(destKey string, key string) *redis.IntCmd {
	return GetRedis().BitOpNot(context.Background(), GetCacheKey(destKey), GetCacheKey(key))
}

// ============================================================================
// 发布订阅 - Publish/Subscribe Operations
// ============================================================================

// Publish publishes message to channel
func Publish(channel string, message interface{}) *redis.IntCmd {
	return GetRedis().Publish(context.Background(), GetCacheKey(channel), message)
}

// Subscribe subscribes to channels
func Subscribe(channels ...string) *redis.PubSub {
	ctx := context.Background()
	newChannels := make([]string, len(channels))
	for i, channel := range channels {
		newChannels[i] = GetCacheKey(channel)
	}
	return GetRedis().Subscribe(ctx, newChannels...)
}

// PSubscribe subscribes to channels matching patterns
func PSubscribe(patterns ...string) *redis.PubSub {
	ctx := context.Background()
	newPatterns := make([]string, len(patterns))
	for i, pattern := range patterns {
		newPatterns[i] = GetCacheKey(pattern)
	}
	return GetRedis().PSubscribe(ctx, newPatterns...)
}

// ============================================================================
// 事务和管道 - Transaction and Pipeline
// ============================================================================

// Pipeline creates a new pipeline
func Pipeline() redis.Pipeliner {
	return GetRedis().Pipeline()
}

// TxPipeline creates a new transaction pipeline
func TxPipeline() redis.Pipeliner {
	return GetRedis().TxPipeline()
}

// ============================================================================
// 便捷缓存操作 - Convenient Cache Operations
// ============================================================================

// GetOrSet gets value by key, if not exists, set it using the setter function
func GetOrSet(key string, expiration time.Duration, setter func() (interface{}, error)) (string, error) {
	ctx := context.Background()
	cacheKey := GetCacheKey(key)

	// Try to get value first
	val, err := GetRedis().Get(ctx, cacheKey).Result()
	if err == nil {
		return val, nil
	}

	if err != redis.Nil {
		return "", err
	}

	// Value doesn't exist, call setter
	newVal, err := setter()
	if err != nil {
		return "", err
	}

	// Set the new value
	err = GetRedis().Set(ctx, cacheKey, newVal, expiration).Err()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", newVal), nil
}

// GetOrSetJSON gets JSON value by key, if not exists, set it using the setter function
func GetOrSetJSON(key string, expiration time.Duration, result interface{}, setter func() (interface{}, error)) error {
	ctx := context.Background()
	cacheKey := GetCacheKey(key)

	// Try to get value first
	val, err := GetRedis().Get(ctx, cacheKey).Result()
	if err == nil {
		return json.Unmarshal([]byte(val), result)
	}

	if err != redis.Nil {
		return err
	}

	// Value doesn't exist, call setter
	newVal, err := setter()
	if err != nil {
		return err
	}

	// Set the new value as JSON
	jsonData, err := json.Marshal(newVal)
	if err != nil {
		return err
	}

	err = GetRedis().Set(ctx, cacheKey, jsonData, expiration).Err()
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, result)
}

// RememberForever caches result forever until manually deleted
func RememberForever(key string, setter func() (interface{}, error)) (string, error) {
	return GetOrSet(key, 0, setter)
}

// RememberForeverJSON caches JSON result forever until manually deleted
func RememberForeverJSON(key string, result interface{}, setter func() (interface{}, error)) error {
	return GetOrSetJSON(key, 0, result, setter)
}

// FlushAll flushes all keys (use with caution)
func FlushAll() *redis.StatusCmd {
	return GetRedis().FlushAll(context.Background())
}

// FlushDB flushes current database
func FlushDB() *redis.StatusCmd {
	return GetRedis().FlushDB(context.Background())
}

// ============================================================================
// 分布式锁 - Distributed Lock
// ============================================================================

// Lock acquires a distributed lock
func Lock(key string, expiration time.Duration, value string) *redis.BoolCmd {
	return GetRedis().SetNX(context.Background(), GetCacheKey(fmt.Sprintf("lock:%s", key)), value, expiration)
}

// Unlock releases a distributed lock
func Unlock(key string, value string) error {
	ctx := context.Background()
	lockKey := GetCacheKey(fmt.Sprintf("lock:%s", key))

	// Lua script to safely unlock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	return GetRedis().Eval(ctx, script, []string{lockKey}, value).Err()
}

// ============================================================================
// 速率限制 - Rate Limiting
// ============================================================================

// RateLimit implements sliding window rate limiting
func RateLimit(key string, limit int, window time.Duration) (bool, error) {
	ctx := context.Background()
	rateLimitKey := GetCacheKey(fmt.Sprintf("rate_limit:%s", key))

	// Lua script for sliding window rate limiting
	script := `
		local key = KEYS[1]
		local window = tonumber(ARGV[1])
		local limit = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		-- Remove expired entries
		redis.call('zremrangebyscore', key, 0, now - window)
		
		-- Count current requests
		local current = redis.call('zcard', key)
		
		if current < limit then
			-- Add current request
			redis.call('zadd', key, now, now)
			redis.call('expire', key, window)
			return 1
		else
			return 0
		end
	`

	result, err := GetRedis().Eval(ctx, script, []string{rateLimitKey},
		int64(window.Seconds()), limit, time.Now().UnixNano()).Int()

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// ============================================================================
// 布隆过滤器相关 - Bloom Filter Related
// ============================================================================

// BFAdd adds item to bloom filter (requires Redis with bloom filter module)
func BFAdd(key string, item string) *redis.Cmd {
	return GetRedis().Do(context.Background(), "BF.ADD", GetCacheKey(key), item)
}

// BFExists checks if item exists in bloom filter
func BFExists(key string, item string) *redis.Cmd {
	return GetRedis().Do(context.Background(), "BF.EXISTS", GetCacheKey(key), item)
}

// ============================================================================
// 批量操作辅助函数 - Batch Operation Helpers
// ============================================================================

// BatchSet sets multiple keys with same expiration
func BatchSet(expiration time.Duration, keyValues ...interface{}) error {
	if len(keyValues)%2 != 0 {
		return fmt.Errorf("keyValues must be even number of arguments")
	}

	ctx := context.Background()
	pipe := GetRedis().Pipeline()

	for i := 0; i < len(keyValues); i += 2 {
		key := keyValues[i].(string)
		value := keyValues[i+1]
		pipe.Set(ctx, GetCacheKey(key), value, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BatchGet gets multiple keys and returns map
func BatchGet(keys ...string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	ctx := context.Background()
	pipe := GetRedis().Pipeline()

	// Prepare commands
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, GetCacheKey(key))
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Collect results
	result := make(map[string]string)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			result[keys[i]] = val
		}
	}

	return result, nil
}

// ============================================================================
// 健康检查 - Health Check
// ============================================================================

// Ping checks if Redis is available
func Ping() *redis.StatusCmd {
	return GetRedis().Ping(context.Background())
}

// Info returns Redis server information
func Info(section ...string) *redis.StringCmd {
	return GetRedis().Info(context.Background(), section...)
}

// ============================================================================
// 数据库选择 - Database Selection
// ============================================================================

// Select changes the database for the connection
func Select(db int) *redis.Cmd {
	return GetRedis().Do(context.Background(), "SELECT", db)
}
