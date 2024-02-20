package helper

import (
	"fmt"
	"strconv"
	"strings"
)

// NiceNumber is nice number
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
		return fmt.Sprintf("%.1f T", float64(n)*0.000000000001)
	}
	if n > 1000000000 {
		return fmt.Sprintf("%.1f B", float64(n)*0.000000001)
	}
	if n > 1000000 {
		return fmt.Sprintf("%.1f M", float64(n)*0.000001)
	}
	if n > 1000 {
		return fmt.Sprintf("%.1f K", float64(n)*0.001)
	}
	return fmt.Sprintf("%d", n)
}

// Number2Icon is number to icon
func Number2Icon(number int) string {
	result := []string{}
	numbers := []string{"0️⃣", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	if number >= 0 && number <= 10 {
		return numbers[number]
	}
	textRunes := []rune(strconv.Itoa(number))
	for _, textRune := range textRunes {
		n, _ := strconv.Atoi(string(textRune))
		result = append(result, numbers[n])
	}
	return strings.Join(result, "")
}

// TrimLastZero is trim last zero
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

// abs
func Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
