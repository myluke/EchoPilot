package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsValidEmail 验证邮箱地址格式
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsObjectID is object id
func IsObjectID(v string) bool {
	_, err := primitive.ObjectIDFromHex(v)
	return err == nil
}

// 检测是否为英文
func IsEnglish(str string) bool {
	for _, ch := range str {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || unicode.IsNumber(ch) || unicode.IsSpace(ch) || unicode.IsPunct(ch)) {
			return false
		}
	}
	return true
}

// SignHash creates a hash for signing a message.
func SignHash(data []byte) []byte {
	return crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)))
}

// EnsureOwner verifies that the signature corresponds to the given address.
func EnsureOwner(address, message, signature string) (common.Address, error) {
	address1 := common.HexToAddress(address)
	rawSig := common.FromHex(signature)

	if len(rawSig) != 65 { // Ethereum signatures are 65 bytes
		return common.Address{}, errors.New("bad signature length")
	}

	rawSig[64] -= 27 // Adjust the recovery ID

	publicKey, err := crypto.SigToPub(SignHash([]byte(message)), rawSig)
	if err != nil {
		return common.Address{}, err
	}

	if owner := crypto.PubkeyToAddress(*publicKey); owner != address1 {
		return common.Address{}, errors.New("mismatch")
	}

	return address1, nil
}

// ==================== 新增的验证函数 ====================

// IsValidPhone 验证手机号码（简单验证）
func IsValidPhone(phone string) bool {
	// 支持国际格式 +1234567890 或 1234567890
	pattern := `^(\+?[1-9]\d{1,14})$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsChinesePhone 验证中国手机号码
func IsChinesePhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsIDCard 验证身份证号码（18位）
func IsIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}

	// 验证格式
	pattern := `^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`
	matched, _ := regexp.MatchString(pattern, idCard)
	if !matched {
		return false
	}

	// 验证校验码
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	sum := 0
	for i := 0; i < 17; i++ {
		digit := int(idCard[i] - '0')
		sum += digit * weights[i]
	}

	remainder := sum % 11
	expectedCheck := checkCodes[remainder]
	actualCheck := string(idCard[17])

	return strings.ToUpper(actualCheck) == expectedCheck
}

// IsPassword 验证密码强度
func IsPassword(password string, minLength int) bool {
	if len(password) < minLength {
		return false
	}

	// 至少包含一个大写字母、一个小写字母、一个数字
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)

	return hasUpper && hasLower && hasDigit
}

// IsStrongPassword 验证强密码
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	// 至少包含一个大写字母、一个小写字母、一个数字、一个特殊字符
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsUsername 验证用户名
func IsUsername(username string) bool {
	// 3-20个字符，只能包含字母、数字、下划线
	pattern := `^[a-zA-Z0-9_]{3,20}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// IsNumeric 验证是否为数字
func IsNumeric(str string) bool {
	pattern := `^-?\d+(\.\d+)?$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsInteger 验证是否为整数
func IsInteger(str string) bool {
	pattern := `^-?\d+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsPositiveInteger 验证是否为正整数
func IsPositiveInteger(str string) bool {
	pattern := `^[1-9]\d*$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsAlpha 验证是否只包含字母
func IsAlpha(str string) bool {
	pattern := `^[a-zA-Z]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsAlphaNumeric 验证是否只包含字母和数字
func IsAlphaNumeric(str string) bool {
	pattern := `^[a-zA-Z0-9]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsChinese 验证是否为中文
func IsChinese(str string) bool {
	pattern := `^[\u4e00-\u9fa5]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsJSON 验证是否为有效的JSON
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsBase64 验证是否为Base64编码
func IsBase64(str string) bool {
	pattern := `^[A-Za-z0-9+/]*={0,2}$`
	matched, _ := regexp.MatchString(pattern, str)
	if !matched {
		return false
	}

	// 检查长度是否为4的倍数
	return len(str)%4 == 0
}

// IsHex 验证是否为十六进制字符串
func IsHex(str string) bool {
	pattern := `^[0-9a-fA-F]+$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// IsUUID 验证是否为UUID
func IsUUID(str string) bool {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(str))
	return matched
}

// IsCreditCard 验证信用卡号码（简单验证）
func IsCreditCard(cardNumber string) bool {
	// 移除空格和连字符
	cardNumber = regexp.MustCompile(`[\s-]`).ReplaceAllString(cardNumber, "")

	// 检查长度和格式
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return false
	}

	if !regexp.MustCompile(`^\d+$`).MatchString(cardNumber) {
		return false
	}

	// Luhn算法验证
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsLatitude 验证纬度
func IsLatitude(lat string) bool {
	pattern := `^[-+]?([1-8]?\d(\.\d+)?|90(\.0+)?)$`
	matched, _ := regexp.MatchString(pattern, lat)
	return matched
}

// IsLongitude 验证经度
func IsLongitude(lng string) bool {
	pattern := `^[-+]?((1[0-7]|[1-9])?\d(\.\d+)?|180(\.0+)?)$`
	matched, _ := regexp.MatchString(pattern, lng)
	return matched
}

// IsColor 验证颜色值（支持hex, rgb, rgba）
func IsColor(color string) bool {
	patterns := []string{
		`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`,                                     // hex
		`^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$`,                     // rgb
		`^rgba\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*[01](\.\d+)?\s*\)$`, // rgba
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, color)
		if matched {
			return true
		}
	}

	return false
}

// IsDateFormat 验证日期格式
func IsDateFormat(date, format string) bool {
	_, err := time.Parse(format, date)
	return err == nil
}

// IsCommonDate 验证常见日期格式
func IsCommonDate(date string) bool {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if IsDateFormat(date, format) {
			return true
		}
	}

	return false
}

// IsTime 验证时间格式
func IsTime(timeStr string) bool {
	pattern := `^([01]?\d|2[0-3]):[0-5]\d(:[0-5]\d)?$`
	matched, _ := regexp.MatchString(pattern, timeStr)
	return matched
}

// IsRange 验证数值是否在指定范围内
func IsRange(value, min, max float64) bool {
	return value >= min && value <= max
}

// IsLength 验证字符串长度是否在指定范围内
func IsLength(str string, min, max int) bool {
	length := len([]rune(str))
	return length >= min && length <= max
}

// IsIn 验证值是否在指定列表中
func IsIn(value string, list []string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

// IsNotIn 验证值是否不在指定列表中
func IsNotIn(value string, list []string) bool {
	return !IsIn(value, list)
}
