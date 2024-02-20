package helper

import (
	"fmt"
	"testing"
)

func TestValueInSlice(t *testing.T) {
	// 测试字符串类型
	strSlice := []string{"apple", "banana", "cherry"}
	fmt.Println(ValueInSlice("banana", strSlice)) // 输出: true
	fmt.Println(ValueInSlice("grape", strSlice))  // 输出: false

	// 测试整数类型
	intSlice := []int{1, 2, 3, 4, 5}
	fmt.Println(ValueInSlice(3, intSlice)) // 输出: true
	fmt.Println(ValueInSlice(6, intSlice)) // 输出: false

	// 测试其他可比较类型
}