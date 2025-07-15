package helper

import (
	"reflect"
	"sort"
)

// ==================== 基础操作 ====================

// ValueInSlice 检查值是否存在于切片中
func ValueInSlice[T comparable](v T, slice []T) bool {
	for _, s := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// IndexOf 返回元素在切片中的索引，如果不存在返回-1
func IndexOf[T comparable](slice []T, value T) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// LastIndexOf 返回元素在切片中的最后一个索引，如果不存在返回-1
func LastIndexOf[T comparable](slice []T, value T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

// Contains 检查切片是否包含指定元素
func Contains[T comparable](slice []T, value T) bool {
	return IndexOf(slice, value) != -1
}

// SliceContainsAll 检查切片是否包含所有指定元素
func SliceContainsAll[T comparable](slice []T, values ...T) bool {
	for _, value := range values {
		if !Contains(slice, value) {
			return false
		}
	}
	return true
}

// SliceContainsAny 检查切片是否包含任意指定元素
func SliceContainsAny[T comparable](slice []T, values ...T) bool {
	for _, value := range values {
		if Contains(slice, value) {
			return true
		}
	}
	return false
}

// ==================== 添加和移除 ====================

// Append 安全地添加元素到切片
func Append[T any](slice []T, elements ...T) []T {
	return append(slice, elements...)
}

// Prepend 在切片开头添加元素
func Prepend[T any](slice []T, elements ...T) []T {
	return append(elements, slice...)
}

// Insert 在指定位置插入元素
func Insert[T any](slice []T, index int, elements ...T) []T {
	if index < 0 || index > len(slice) {
		return slice
	}

	// 创建足够大的切片
	result := make([]T, len(slice)+len(elements))

	// 复制前半部分
	copy(result[:index], slice[:index])

	// 插入新元素
	copy(result[index:index+len(elements)], elements)

	// 复制后半部分
	copy(result[index+len(elements):], slice[index:])

	return result
}

// Remove 移除指定索引的元素
func Remove[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}

// RemoveValue 移除第一个匹配的元素
func RemoveValue[T comparable](slice []T, value T) []T {
	index := IndexOf(slice, value)
	if index != -1 {
		return Remove(slice, index)
	}
	return slice
}

// RemoveAll 移除所有匹配的元素
func RemoveAll[T comparable](slice []T, value T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

// RemoveByIndices 移除多个索引的元素
func RemoveByIndices[T any](slice []T, indices ...int) []T {
	if len(indices) == 0 {
		return slice
	}

	// 排序索引并去重
	sort.Ints(indices)
	unique := make([]int, 0, len(indices))
	for i, idx := range indices {
		if i == 0 || idx != indices[i-1] {
			if idx >= 0 && idx < len(slice) {
				unique = append(unique, idx)
			}
		}
	}

	if len(unique) == 0 {
		return slice
	}

	// 从后往前删除
	result := make([]T, len(slice))
	copy(result, slice)

	for i := len(unique) - 1; i >= 0; i-- {
		idx := unique[i]
		result = append(result[:idx], result[idx+1:]...)
	}

	return result
}

// ==================== 切片操作 ====================

// Slice 安全地切片
func Slice[T any](slice []T, start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(slice) {
		end = len(slice)
	}
	if start >= end {
		return []T{}
	}

	return slice[start:end]
}

// SubSlice 获取子切片
func SubSlice[T any](slice []T, start int, length int) []T {
	if start < 0 {
		start = 0
	}
	if start >= len(slice) {
		return []T{}
	}

	end := start + length
	if end > len(slice) {
		end = len(slice)
	}

	return slice[start:end]
}

// Head 获取切片的前n个元素
func Head[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return append([]T{}, slice...)
	}

	return slice[:n]
}

// Tail 获取切片的后n个元素
func Tail[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return append([]T{}, slice...)
	}

	return slice[len(slice)-n:]
}

// Drop 丢弃前n个元素
func Drop[T any](slice []T, n int) []T {
	if n <= 0 {
		return append([]T{}, slice...)
	}
	if n >= len(slice) {
		return []T{}
	}

	return slice[n:]
}

// DropRight 丢弃后n个元素
func DropRight[T any](slice []T, n int) []T {
	if n <= 0 {
		return append([]T{}, slice...)
	}
	if n >= len(slice) {
		return []T{}
	}

	return slice[:len(slice)-n]
}

// ==================== 变换操作 ====================

// Map 映射操作
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter 过滤操作
func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce 归约操作
func Reduce[T, U any](slice []T, fn func(U, T) U, initial U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// ForEach 遍历操作
func ForEach[T any](slice []T, fn func(T)) {
	for _, v := range slice {
		fn(v)
	}
}

// Any 检查是否存在满足条件的元素
func Any[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if fn(v) {
			return true
		}
	}
	return false
}

// All 检查是否所有元素都满足条件
func All[T any](slice []T, fn func(T) bool) bool {
	for _, v := range slice {
		if !fn(v) {
			return false
		}
	}
	return true
}

// Find 查找第一个满足条件的元素
func Find[T any](slice []T, fn func(T) bool) (T, bool) {
	for _, v := range slice {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex 查找第一个满足条件的元素索引
func FindIndex[T any](slice []T, fn func(T) bool) int {
	for i, v := range slice {
		if fn(v) {
			return i
		}
	}
	return -1
}

// FindLast 查找最后一个满足条件的元素
func FindLast[T any](slice []T, fn func(T) bool) (T, bool) {
	for i := len(slice) - 1; i >= 0; i-- {
		if fn(slice[i]) {
			return slice[i], true
		}
	}
	var zero T
	return zero, false
}

// FindLastIndex 查找最后一个满足条件的元素索引
func FindLastIndex[T any](slice []T, fn func(T) bool) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if fn(slice[i]) {
			return i
		}
	}
	return -1
}

// ==================== 排序和去重 ====================

// Unique 去重
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// UniqueBy 根据函数去重
func UniqueBy[T any, K comparable](slice []T, fn func(T) K) []T {
	seen := make(map[K]bool)
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		key := fn(v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	return result
}

// SliceReverse 反转切片
func SliceReverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// Sort 排序（需要实现排序接口）
func Sort[T any](slice []T, less func(T, T) bool) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	sort.Slice(result, func(i, j int) bool {
		return less(result[i], result[j])
	})

	return result
}

// ==================== 集合操作 ====================

// Union 并集
func Union[T comparable](slice1, slice2 []T) []T {
	result := make([]T, 0, len(slice1)+len(slice2))
	seen := make(map[T]bool)

	for _, v := range slice1 {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	for _, v := range slice2 {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// Intersection 交集
func Intersection[T comparable](slice1, slice2 []T) []T {
	set1 := make(map[T]bool)
	for _, v := range slice1 {
		set1[v] = true
	}

	result := make([]T, 0)
	seen := make(map[T]bool)

	for _, v := range slice2 {
		if set1[v] && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// Difference 差集 (slice1 - slice2)
func Difference[T comparable](slice1, slice2 []T) []T {
	set2 := make(map[T]bool)
	for _, v := range slice2 {
		set2[v] = true
	}

	result := make([]T, 0)
	for _, v := range slice1 {
		if !set2[v] {
			result = append(result, v)
		}
	}

	return result
}

// SymmetricDifference 对称差集
func SymmetricDifference[T comparable](slice1, slice2 []T) []T {
	set1 := make(map[T]bool)
	for _, v := range slice1 {
		set1[v] = true
	}

	set2 := make(map[T]bool)
	for _, v := range slice2 {
		set2[v] = true
	}

	result := make([]T, 0)

	for _, v := range slice1 {
		if !set2[v] {
			result = append(result, v)
		}
	}

	for _, v := range slice2 {
		if !set1[v] {
			result = append(result, v)
		}
	}

	return result
}

// ==================== 分组和分割 ====================

// GroupBy 按条件分组
func GroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)

	for _, v := range slice {
		key := fn(v)
		result[key] = append(result[key], v)
	}

	return result
}

// Partition 分割为两部分
func Partition[T any](slice []T, fn func(T) bool) ([]T, []T) {
	var trueSlice, falseSlice []T

	for _, v := range slice {
		if fn(v) {
			trueSlice = append(trueSlice, v)
		} else {
			falseSlice = append(falseSlice, v)
		}
	}

	return trueSlice, falseSlice
}

// Chunk 分块
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	var result [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}

// ==================== 统计和比较 ====================

// Count 计数
func Count[T comparable](slice []T, value T) int {
	count := 0
	for _, v := range slice {
		if v == value {
			count++
		}
	}
	return count
}

// CountBy 按条件计数
func CountBy[T any](slice []T, fn func(T) bool) int {
	count := 0
	for _, v := range slice {
		if fn(v) {
			count++
		}
	}
	return count
}

// Equal 比较两个切片是否相等
func Equal[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}

	return true
}

// DeepEqual 深度比较两个切片
func DeepEqual(slice1, slice2 interface{}) bool {
	return reflect.DeepEqual(slice1, slice2)
}

// ==================== 其他工具 ====================

// SliceIsEmpty 检查切片是否为空
func SliceIsEmpty[T any](slice []T) bool {
	return len(slice) == 0
}

// IsNotEmpty 检查切片是否非空
func IsNotEmpty[T any](slice []T) bool {
	return len(slice) > 0
}

// Clone 克隆切片
func Clone[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	return result
}

// Concat 连接多个切片
func Concat[T any](slices ...[]T) []T {
	totalLen := 0
	for _, slice := range slices {
		totalLen += len(slice)
	}

	result := make([]T, 0, totalLen)
	for _, slice := range slices {
		result = append(result, slice...)
	}

	return result
}

// Flatten 展开二维切片
func Flatten[T any](slices [][]T) []T {
	totalLen := 0
	for _, slice := range slices {
		totalLen += len(slice)
	}

	result := make([]T, 0, totalLen)
	for _, slice := range slices {
		result = append(result, slice...)
	}

	return result
}

// Zip 压缩两个切片
func Zip[T, U any](slice1 []T, slice2 []U) []struct {
	First  T
	Second U
} {
	length := len(slice1)
	if len(slice2) < length {
		length = len(slice2)
	}

	result := make([]struct {
		First  T
		Second U
	}, length)
	for i := 0; i < length; i++ {
		result[i] = struct {
			First  T
			Second U
		}{slice1[i], slice2[i]}
	}

	return result
}

// Unzip 解压缩切片
func Unzip[T, U any](slice []struct {
	First  T
	Second U
}) ([]T, []U) {
	slice1 := make([]T, len(slice))
	slice2 := make([]U, len(slice))

	for i, v := range slice {
		slice1[i] = v.First
		slice2[i] = v.Second
	}

	return slice1, slice2
}

// Sample 随机采样
func Sample[T any](slice []T, n int) []T {
	if n >= len(slice) {
		return Clone(slice)
	}

	// 使用 Fisher-Yates 洗牌算法
	result := Clone(slice)
	for i := 0; i < n; i++ {
		j := RandRange(i, len(result)-1)
		result[i], result[j] = result[j], result[i]
	}

	return result[:n]
}

// Shuffle 随机打乱
func Shuffle[T any](slice []T) []T {
	result := Clone(slice)
	for i := len(result) - 1; i > 0; i-- {
		j := RandRange(0, i)
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Repeat 重复元素
func Repeat[T any](value T, count int) []T {
	result := make([]T, count)
	for i := 0; i < count; i++ {
		result[i] = value
	}
	return result
}

// Range 生成数字范围
func Range(start, end int) []int {
	if start >= end {
		return []int{}
	}

	result := make([]int, end-start)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// RangeStep 生成带步长的数字范围
func RangeStep(start, end, step int) []int {
	if step == 0 || (step > 0 && start >= end) || (step < 0 && start <= end) {
		return []int{}
	}

	result := make([]int, 0)
	if step > 0 {
		for i := start; i < end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i > end; i += step {
			result = append(result, i)
		}
	}

	return result
}
