package helper

import (
	"fmt"
	"hash/fnv"
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

// 通用的转换函数
func convert[T Convertible](v T) string {
	switch value := any(v).(type) {
	case string:
		return value
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(value).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case primitive.ObjectID:
		return value.Hex()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// ToInt64 泛型函数，尝试将不同的类型转换为int64。
func ToInt64[T Convertible](v T) int64 {
	str := convert(v)
	r, _ := strconv.ParseInt(str, 10, 64)
	return r
}

// ToFloat64 泛型函数，尝试将不同的类型转换为float64。
func ToFloat64[T Convertible](v T) float64 {
	str := convert(v)
	r, _ := strconv.ParseFloat(str, 64)
	return r
}

// ToObjectID 泛型函数，尝试将不同的类型转换为ObjectID。
func ToObjectID[T Convertible](v T) primitive.ObjectID {
	str := convert(v)
	r, _ := primitive.ObjectIDFromHex(str)
	return r
}

// ToString 泛型函数，尝试将不同的类型转换为string。
func ToString[T Convertible](v T) string {
	return convert(v)
}

// ToUInt32 泛型函数，尝试将不同的类型转换为uint32。
func ToUInt32[T Convertible](v T) uint32 {
	str := convert(v)
	r, _ := strconv.ParseUint(str, 10, 32)
	return uint32(r)
}

// ToUInt64 泛型函数，尝试将不同的类型转换为uint64。
func ToUInt64[T Convertible](v T) uint64 {
	str := convert(v)
	r, _ := strconv.ParseUint(str, 10, 64)
	return r
}

// ToFNV32Hash 泛型函数，使用 FNV-1a 算法将输入转换为 32 位哈希值。
func ToFNV32Hash[T Convertible](v T) uint32 {
	str := convert(v)
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}

// ToFNV64Hash 泛型函数，使用 FNV-1a 算法将输入转换为 64 位哈希值。
func ToFNV64Hash[T Convertible](v T) uint64 {
	str := convert(v)
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}
