package helper

import (
	"fmt"
	"hash/fnv"
	"math"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 定义一个通用的类型约束
type Convertible interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~bool | primitive.ObjectID
}

// 修改后的 convert 函数
func convert[T Convertible](v T) interface{} {
	switch value := any(v).(type) {
	case string:
		return value
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value).Uint()
	case float32, float64:
		return reflect.ValueOf(value).Float()
	case bool:
		return value
	case primitive.ObjectID:
		return value
	default:
		return fmt.Sprintf("%v", value)
	}
}

// 修改 ToInt64 函数
func ToInt64[T Convertible](v T) int64 {
	switch value := convert(v).(type) {
	case int64:
		return value
	case uint64:
		if value <= math.MaxInt64 {
			return int64(value)
		}
		return math.MaxInt64
	case float64:
		return int64(math.Round(value))
	case string:
		r, _ := strconv.ParseInt(value, 10, 64)
		return r
	default:
		return 0
	}
}

// 修改 ToFloat64 函数
func ToFloat64[T Convertible](v T) float64 {
	switch value := convert(v).(type) {
	case float64:
		return value
	case int64:
		return float64(value)
	case uint64:
		return float64(value)
	case string:
		r, _ := strconv.ParseFloat(value, 64)
		return r
	default:
		return 0
	}
}

// ToObjectID 泛型函数，尝试将不同的类型转换为ObjectID。
func ToObjectID[T Convertible](v T) primitive.ObjectID {
	str := ToString(v)
	r, _ := primitive.ObjectIDFromHex(str)
	return r
}

// ToString 泛型函数，尝试将不同的类型转换为string。
func ToString[T Convertible](v T) string {
	switch value := convert(v).(type) {
	case string:
		return value
	case int64:
		return strconv.FormatInt(value, 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case primitive.ObjectID:
		return value.Hex()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// ToUInt32 泛型函数，尝试将不同的类型转换为uint32。
func ToUInt32[T Convertible](v T) uint32 {
	switch value := convert(v).(type) {
	case uint64:
		if value <= math.MaxUint32 {
			return uint32(value)
		}
		return math.MaxUint32
	case int64:
		if value >= 0 && value <= math.MaxUint32 {
			return uint32(value)
		}
		return 0
	case float64:
		if value >= 0 && value <= math.MaxUint32 {
			return uint32(math.Round(value))
		}
		return 0
	case string:
		r, _ := strconv.ParseUint(value, 10, 32)
		return uint32(r)
	default:
		return 0
	}
}

// ToUInt64 泛型函数，尝试将不同的类型转换为uint64。
func ToUInt64[T Convertible](v T) uint64 {
	switch value := convert(v).(type) {
	case uint64:
		return value
	case int64:
		if value >= 0 {
			return uint64(value)
		}
		return 0
	case float64:
		if value >= 0 {
			return uint64(math.Round(value))
		}
		return 0
	case string:
		r, _ := strconv.ParseUint(value, 10, 64)
		return r
	default:
		return 0
	}
}

// ToFNV32Hash 泛型函数，使用 FNV-1a 算法将输入转换为 32 位哈希值。
func ToFNV32Hash[T Convertible](v T) uint32 {
	str := ToString(v)
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}

// ToFNV64Hash 泛型函数，使用 FNV-1a 算法将输入转换为 64 位哈希值。
func ToFNV64Hash[T Convertible](v T) uint64 {
	str := ToString(v)
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}

// 修改 ToInt32 函数
func ToInt32[T Convertible](v T) int32 {
	switch value := convert(v).(type) {
	case int64:
		if value >= math.MinInt32 && value <= math.MaxInt32 {
			return int32(value)
		}
		if value > math.MaxInt32 {
			return math.MaxInt32
		}
		return math.MinInt32
	case uint64:
		if value <= math.MaxInt32 {
			return int32(value)
		}
		return math.MaxInt32
	case float64:
		if value >= math.MinInt32 && value <= math.MaxInt32 {
			return int32(math.Round(value))
		}
		if value > math.MaxInt32 {
			return math.MaxInt32
		}
		return math.MinInt32
	case string:
		r, _ := strconv.ParseInt(value, 10, 32)
		return int32(r)
	default:
		return 0
	}
}

// 修改 ToBool 函数
func ToBool[T Convertible](v T) bool {
	switch value := convert(v).(type) {
	case bool:
		return value
	case int64:
		return value != 0
	case uint64:
		return value != 0
	case float64:
		return value != 0
	case string:
		r, _ := strconv.ParseBool(value)
		return r
	default:
		return false
	}
}
