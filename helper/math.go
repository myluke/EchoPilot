package helper

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ==================== 数字格式化 ====================

// NiceNumber 格式化数字为友好显示
func NiceNumber(num any) string {
	var n int64 = 0
	switch _n := num.(type) {
	case int64:
		n = _n
	case int32:
		n = int64(_n)
	case int:
		n = int64(_n)
	case string:
		i, _ := strconv.Atoi(_n)
		n = int64(i)
	}

	if n > 1000000000000 {
		return TrimLastZero(float64(n)*0.000000000001, "%.1f") + " T"
	}
	if n > 1000000000 {
		return TrimLastZero(float64(n)*0.000000001, "%.1f") + " B"
	}
	if n > 1000000 {
		return TrimLastZero(float64(n)*0.000001, "%.1f") + " M"
	}
	if n > 1000 {
		return TrimLastZero(float64(n)*0.001, "%.1f") + " K"
	}
	return fmt.Sprintf("%d", n)
}

// Number2Icon 数字转图标
func Number2Icon(number int) string {
	numbers := []string{"0️⃣", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	if number >= 0 && number <= 10 {
		return numbers[number]
	}

	result := make([]string, 0)
	textRunes := []rune(strconv.Itoa(number))
	for _, textRune := range textRunes {
		n, _ := strconv.Atoi(string(textRune))
		result = append(result, numbers[n])
	}
	return strings.Join(result, "")
}

// TrimLastZero 去掉末尾零
func TrimLastZero(f float64, p ...string) string {
	s := fmt.Sprintf("%.6f", f)
	if len(p) > 0 {
		s = fmt.Sprintf(p[0], f)
	}

	if !strings.Contains(s, ".") {
		return s
	}

	args := strings.Split(s, ".")
	args[1] = strings.TrimRight(args[1], "0")
	if len(args[1]) > 0 {
		return fmt.Sprintf("%s.%s", args[0], args[1])
	}
	return args[0]
}

// ==================== 基础数学运算 ====================

// Abs 绝对值
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

// AbsFloat 浮点数绝对值
func AbsFloat(f float64) float64 {
	return math.Abs(f)
}

// Max 返回两个整数中的最大值
func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Min 返回两个整数中的最小值
func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxFloat 返回两个浮点数中的最大值
func MaxFloat(a, b float64) float64 {
	return math.Max(a, b)
}

// MinFloat 返回两个浮点数中的最小值
func MinFloat(a, b float64) float64 {
	return math.Min(a, b)
}

// Clamp 限制值在指定范围内
func Clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampFloat 限制浮点数在指定范围内
func ClampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ==================== 随机数生成 ====================

// RandRange 生成指定范围的随机整数
func RandRange(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

// RandFloatRange 生成指定范围的随机浮点数
func RandFloatRange(min, max float64) float64 {
	if min >= max {
		return min
	}

	value := rand.Float64()*(max-min) + min
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// RandDoubleAverage 二倍均值算法生成随机数
func RandDoubleAverage(count int64, min, max float64) float64 {
	if count == 1 {
		return max
	}

	avg := max / float64(count)
	avg2 := 2*avg + min
	value := rand.Float64()*(avg2) + min
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// RandString 生成随机字符串
func RandString(n int) string {
	const letterRunes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]rune, n)
	for i := range b {
		b[i] = rune(letterRunes[rand.Intn(len(letterRunes))])
	}
	return string(b)
}

// RandBytes 生成随机字节切片
func RandBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

// RandBool 生成随机布尔值
func RandBool() bool {
	return rand.Intn(2) == 1
}

// RandChoice 从切片中随机选择一个元素
func RandChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rand.Intn(len(slice))]
}

// RandShuffle 随机打乱切片
func RandShuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// RandSample 从切片中随机抽取 n 个元素
func RandSample[T any](slice []T, n int) []T {
	if n >= len(slice) {
		return RandShuffle(slice)
	}

	shuffled := RandShuffle(slice)
	return shuffled[:n]
}

// ==================== 数学计算 ====================

// Sum 计算整数切片的和
func Sum(numbers []int64) int64 {
	var total int64
	for _, num := range numbers {
		total += num
	}
	return total
}

// SumFloat 计算浮点数切片的和
func SumFloat(numbers []float64) float64 {
	var total float64
	for _, num := range numbers {
		total += num
	}
	return total
}

// Average 计算平均值
func Average(numbers []int64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return float64(Sum(numbers)) / float64(len(numbers))
}

// AverageFloat 计算浮点数平均值
func AverageFloat(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return SumFloat(numbers) / float64(len(numbers))
}

// Median 计算中位数
func Median(numbers []int64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	// 复制并排序
	sorted := make([]int64, len(numbers))
	copy(sorted, numbers)

	// 简单冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return float64(sorted[n/2-1]+sorted[n/2]) / 2
	}
	return float64(sorted[n/2])
}

// MedianFloat 计算浮点数中位数
func MedianFloat(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	// 复制并排序
	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)

	// 简单冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// Mode 计算众数
func Mode(numbers []int64) int64 {
	if len(numbers) == 0 {
		return 0
	}

	counts := make(map[int64]int)
	for _, num := range numbers {
		counts[num]++
	}

	var mode int64
	var maxCount int
	for num, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = num
		}
	}

	return mode
}

// StandardDeviation 计算标准差
func StandardDeviation(numbers []int64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	avg := Average(numbers)
	var sum float64
	for _, num := range numbers {
		diff := float64(num) - avg
		sum += diff * diff
	}

	return math.Sqrt(sum / float64(len(numbers)))
}

// StandardDeviationFloat 计算浮点数标准差
func StandardDeviationFloat(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	avg := AverageFloat(numbers)
	var sum float64
	for _, num := range numbers {
		diff := num - avg
		sum += diff * diff
	}

	return math.Sqrt(sum / float64(len(numbers)))
}

// ==================== 数字工具 ====================

// IsPrime 判断是否为质数
func IsPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := int64(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}

	return true
}

// GCD 计算最大公约数
func GCD(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return Abs(a)
}

// LCM 计算最小公倍数
func LCM(a, b int64) int64 {
	return Abs(a*b) / GCD(a, b)
}

// Factorial 计算阶乘
func Factorial(n int64) int64 {
	if n < 0 {
		return 0
	}
	if n == 0 {
		return 1
	}

	result := int64(1)
	for i := int64(2); i <= n; i++ {
		result *= i
	}
	return result
}

// Fibonacci 计算斐波那契数列第n项
func Fibonacci(n int64) int64 {
	if n <= 1 {
		return n
	}

	a, b := int64(0), int64(1)
	for i := int64(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// Power 计算幂次方
func Power(base, exp int64) int64 {
	if exp < 0 {
		return 0
	}
	if exp == 0 {
		return 1
	}

	result := int64(1)
	for i := int64(0); i < exp; i++ {
		result *= base
	}
	return result
}

// PowerFloat 计算浮点数幂次方
func PowerFloat(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Sqrt 计算平方根
func Sqrt(n float64) float64 {
	return math.Sqrt(n)
}

// Cbrt 计算立方根
func Cbrt(n float64) float64 {
	return math.Cbrt(n)
}

// Log 计算自然对数
func Log(n float64) float64 {
	return math.Log(n)
}

// Log10 计算以10为底的对数
func Log10(n float64) float64 {
	return math.Log10(n)
}

// Log2 计算以2为底的对数
func Log2(n float64) float64 {
	return math.Log2(n)
}

// Round 四舍五入
func Round(f float64) float64 {
	return math.Round(f)
}

// RoundToDecimal 四舍五入到指定小数位
func RoundToDecimal(f float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(f*multiplier) / multiplier
}

// Ceil 向上取整
func Ceil(f float64) float64 {
	return math.Ceil(f)
}

// Floor 向下取整
func Floor(f float64) float64 {
	return math.Floor(f)
}

// IsEven 判断是否为偶数
func IsEven(n int64) bool {
	return n%2 == 0
}

// IsOdd 判断是否为奇数
func IsOdd(n int64) bool {
	return n%2 != 0
}

// Sign 获取数字符号
func Sign(n int64) int {
	if n > 0 {
		return 1
	} else if n < 0 {
		return -1
	}
	return 0
}

// SignFloat 获取浮点数符号
func SignFloat(f float64) int {
	if f > 0 {
		return 1
	} else if f < 0 {
		return -1
	}
	return 0
}

// ==================== 进制转换 ====================

// ToBinary 转换为二进制字符串
func ToBinary(n int64) string {
	return strconv.FormatInt(n, 2)
}

// ToOctal 转换为八进制字符串
func ToOctal(n int64) string {
	return strconv.FormatInt(n, 8)
}

// ToHex 转换为十六进制字符串
func ToHex(n int64) string {
	return strconv.FormatInt(n, 16)
}

// FromBinary 从二进制字符串转换
func FromBinary(s string) (int64, error) {
	return strconv.ParseInt(s, 2, 64)
}

// FromOctal 从八进制字符串转换
func FromOctal(s string) (int64, error) {
	return strconv.ParseInt(s, 8, 64)
}

// FromHex 从十六进制字符串转换
func FromHex(s string) (int64, error) {
	return strconv.ParseInt(s, 16, 64)
}

// ==================== 数值范围 ====================

// InRange 检查数字是否在指定范围内
func InRange(value, min, max int64) bool {
	return value >= min && value <= max
}

// InRangeFloat 检查浮点数是否在指定范围内
func InRangeFloat(value, min, max float64) bool {
	return value >= min && value <= max
}

// Normalize 将数字标准化到[0,1]范围
func Normalize(value, min, max float64) float64 {
	if max == min {
		return 0
	}
	return (value - min) / (max - min)
}

// Denormalize 将标准化的数字还原到原始范围
func Denormalize(normalized, min, max float64) float64 {
	return normalized*(max-min) + min
}

// Lerp 线性插值
func Lerp(start, end, t float64) float64 {
	return start + t*(end-start)
}

// InverseLerp 反线性插值
func InverseLerp(start, end, value float64) float64 {
	if end == start {
		return 0
	}
	return (value - start) / (end - start)
}
