package helper

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// RandRange is 随机一个范围
func RandRange(min int, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}

// RandFloatRange
func RandFloatRange(min float64, max float64) float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	value := r.Float64()*(max-min) + min
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// RandDoubleAverage 二倍均值算法生成随机数
func RandDoubleAverage(count int64, min float64, max float64) float64 {
	if count == 1 {
		return max
	}
	avg := max / float64(count)
	avg2 := 2*avg + min
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	value := r.Float64()*(avg2) + min
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// RandString is get rand string
func RandString(n int) string {
	letterRunes := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
