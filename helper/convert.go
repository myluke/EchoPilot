package helper

import (
	"fmt"
	"hash/fnv"
	"math"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Convertible 定义可转换的类型约束
type Convertible interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~bool | primitive.ObjectID
}

// convert 通用转换函数
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

// ToInt64 转换为 int64
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

// ToFloat64 转换为 float64
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

// ToString 转换为字符串
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

// ToInt32 转换为 int32
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

// ToUInt32 转换为 uint32
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

// ToUInt64 转换为 uint64
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

// ToBool 转换为 bool
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

// ToObjectID 转换为 ObjectID
func ToObjectID[T Convertible](v T) primitive.ObjectID {
	str := ToString(v)
	r, _ := primitive.ObjectIDFromHex(str)
	return r
}

// ToFNV32Hash 使用 FNV-1a 算法转换为 32 位哈希值
func ToFNV32Hash[T Convertible](v T) string {
	str := ToString(v)
	h := fnv.New32a()
	h.Write([]byte(str))
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

// ToFNV64Hash 使用 FNV-1a 算法转换为 64 位哈希值
func ToFNV64Hash[T Convertible](v T) string {
	str := ToString(v)
	h := fnv.New64a()
	h.Write([]byte(str))
	return strconv.FormatUint(h.Sum64(), 10)
}

// ==================== 新增的实用函数 ====================

// ToInt 转换为 int
func ToInt[T Convertible](v T) int {
	return int(ToInt64(v))
}

// ToInt8 转换为 int8
func ToInt8[T Convertible](v T) int8 {
	value := ToInt64(v)
	if value >= math.MinInt8 && value <= math.MaxInt8 {
		return int8(value)
	}
	if value > math.MaxInt8 {
		return math.MaxInt8
	}
	return math.MinInt8
}

// ToInt16 转换为 int16
func ToInt16[T Convertible](v T) int16 {
	value := ToInt64(v)
	if value >= math.MinInt16 && value <= math.MaxInt16 {
		return int16(value)
	}
	if value > math.MaxInt16 {
		return math.MaxInt16
	}
	return math.MinInt16
}

// ToUInt 转换为 uint
func ToUInt[T Convertible](v T) uint {
	return uint(ToUInt64(v))
}

// ToUInt8 转换为 uint8
func ToUInt8[T Convertible](v T) uint8 {
	value := ToUInt64(v)
	if value <= math.MaxUint8 {
		return uint8(value)
	}
	return math.MaxUint8
}

// ToUInt16 转换为 uint16
func ToUInt16[T Convertible](v T) uint16 {
	value := ToUInt64(v)
	if value <= math.MaxUint16 {
		return uint16(value)
	}
	return math.MaxUint16
}

// ToFloat32 转换为 float32
func ToFloat32[T Convertible](v T) float32 {
	value := ToFloat64(v)
	if value >= -math.MaxFloat32 && value <= math.MaxFloat32 {
		return float32(value)
	}
	if value > math.MaxFloat32 {
		return math.MaxFloat32
	}
	return -math.MaxFloat32
}

// ToBytes 转换为字节切片
func ToBytes(v any) []byte {
	switch val := v.(type) {
	case []byte:
		return val
	case string:
		return []byte(val)
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf("%d", val))
	case uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf("%d", val))
	case float32, float64:
		return []byte(fmt.Sprintf("%f", val))
	case bool:
		return []byte(fmt.Sprintf("%t", val))
	default:
		return []byte(fmt.Sprintf("%v", val))
	}
}

// FromBytes 从字节切片转换为字符串
func FromBytes(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

// ToRunes 转换为 rune 切片
func ToRunes(s string) []rune {
	return []rune(s)
}

// FromRunes 从 rune 切片转换为字符串
func FromRunes(runes []rune) string {
	return string(runes)
}

// ToTime 转换为时间
func ToTime(v any) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
			"15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("无法解析时间格式: %s", val)
	case int64:
		return time.Unix(val, 0), nil
	case int:
		return time.Unix(int64(val), 0), nil
	case float64:
		return time.Unix(int64(val), 0), nil
	default:
		return time.Time{}, fmt.Errorf("不支持的时间类型: %T", val)
	}
}

// ToTimePtr 转换为时间指针
func ToTimePtr(v any) *time.Time {
	if t, err := ToTime(v); err == nil {
		return &t
	}
	return nil
}

// ToStringPtr 转换为字符串指针
func ToStringPtr[T Convertible](v T) *string {
	str := ToString(v)
	return &str
}

// ToIntPtr 转换为整数指针
func ToIntPtr[T Convertible](v T) *int {
	val := ToInt(v)
	return &val
}

// ToInt64Ptr 转换为 int64 指针
func ToInt64Ptr[T Convertible](v T) *int64 {
	val := ToInt64(v)
	return &val
}

// ToFloat64Ptr 转换为 float64 指针
func ToFloat64Ptr[T Convertible](v T) *float64 {
	val := ToFloat64(v)
	return &val
}

// ToBoolPtr 转换为布尔指针
func ToBoolPtr[T Convertible](v T) *bool {
	val := ToBool(v)
	return &val
}

// SafeToString 安全转换为字符串（处理 nil 指针）
func SafeToString(v any) string {
	if v == nil {
		return ""
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	return fmt.Sprintf("%v", val.Interface())
}

// SafeToInt64 安全转换为 int64（处理 nil 指针）
func SafeToInt64(v any) int64 {
	if v == nil {
		return 0
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(val.Float())
	case reflect.String:
		if i, err := strconv.ParseInt(val.String(), 10, 64); err == nil {
			return i
		}
	}

	return 0
}

// SafeToFloat64 安全转换为 float64（处理 nil 指针）
func SafeToFloat64(v any) float64 {
	if v == nil {
		return 0
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint())
	case reflect.String:
		if f, err := strconv.ParseFloat(val.String(), 64); err == nil {
			return f
		}
	}

	return 0
}

// SafeToBool 安全转换为 bool（处理 nil 指针）
func SafeToBool(v any) bool {
	if v == nil {
		return false
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0
	case reflect.String:
		if b, err := strconv.ParseBool(val.String()); err == nil {
			return b
		}
		return val.String() != ""
	}

	return false
}

// IsNil 检查值是否为 nil
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}

// IsEmpty 检查值是否为空
func IsEmpty(v any) bool {
	if IsNil(v) {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	}

	return false
}

// DeepCopy 深度复制
func DeepCopy(src any) any {
	if src == nil {
		return nil
	}

	val := reflect.ValueOf(src)
	return deepCopyValue(val).Interface()
}

// deepCopyValue 深度复制值
func deepCopyValue(val reflect.Value) reflect.Value {
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		newVal := reflect.New(val.Type().Elem())
		newVal.Elem().Set(deepCopyValue(val.Elem()))
		return newVal

	case reflect.Slice:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		newVal := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newVal.Index(i).Set(deepCopyValue(val.Index(i)))
		}
		return newVal

	case reflect.Map:
		if val.IsNil() {
			return reflect.Zero(val.Type())
		}
		newVal := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newVal.SetMapIndex(key, deepCopyValue(val.MapIndex(key)))
		}
		return newVal

	case reflect.Struct:
		newVal := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).CanSet() {
				newVal.Field(i).Set(deepCopyValue(val.Field(i)))
			}
		}
		return newVal

	default:
		return val
	}
}
