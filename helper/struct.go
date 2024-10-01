package helper

import (
	"reflect"
	"time"
)

// MergeStructs 将 src 结构体的非零值字段合并到 dst 结构体中
func MergeStructs(dst, src interface{}) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src).Elem()

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		dstField := dstValue.Field(i)

		// 检查字段是否可设置且源字段不为零值
		if dstField.CanSet() && !isZeroValue(srcField) {
			dstField.Set(srcField)
		}
	}
}

// Zeroer 是一个接口，用于自定义零值检查
type Zeroer interface {
	IsZero() bool
}

// isZeroValue 检查一个值是否为其类型的零值
func isZeroValue(v reflect.Value) bool {
	// 首先检查是否实现了 Zeroer 接口
	if v.Type().Implements(reflect.TypeOf((*Zeroer)(nil)).Elem()) {
		if v.CanInterface() {
			return v.Interface().(Zeroer).IsZero()
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == complex(0, 0)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		// 特殊处理 time.Time 类型
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// 对于其他结构体，检查所有字段
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	case reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}
