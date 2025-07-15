package helper

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 预编译正则表达式以提高性能
var (
	tagsRegex      = regexp.MustCompile(`[,，、;；\|\s]\s*`)
	specialRegex   = regexp.MustCompile(`[\x00-\x1F]|[\x21-\x2F]|[\x3A-\x40]|[\x5B-\x60]|[\x7B-\x7F]`)
	chineseRegex   = regexp.MustCompile(`[【】、，。？《》～！¥……（）——；'："「」｜]`)
	newlineRegex   = regexp.MustCompile(`\r|\n|\t`)
	nonNumberRegex = regexp.MustCompile(`[^\-\d\.]+`)
	spaceRegex     = regexp.MustCompile(`\s+`)
	htmlSpaceRegex = regexp.MustCompile(`>\s+<`)
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex     = regexp.MustCompile(`^[+]?[\d\s\-\(\)]{7,15}$`)
	ipv4Regex      = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
)

// FindSubstr 查找截取字符串
func FindSubstr(content string, params ...any) string {
	if len(params) < 2 {
		return ""
	}

	startArg := params[0]
	endArg := params[1]

	var startPos int
	switch start := startArg.(type) {
	case string:
		startPos = strings.Index(content, start)
		if startPos > -1 {
			startPos += len(start)
		}
	case *regexp.Regexp:
		startR := start.FindStringIndex(content)
		if startR == nil {
			startPos = -1
		} else {
			startPos = startR[1]
		}
	}

	if startPos == -1 {
		return ""
	}

	remainContent := content[startPos:]

	var endPos int
	switch end := endArg.(type) {
	case string:
		endPos = strings.Index(remainContent, end)
	case *regexp.Regexp:
		endR := end.FindStringIndex(remainContent)
		if endR == nil {
			endPos = -1
		} else {
			endPos = endR[0]
		}
	}

	if endPos == -1 {
		return ""
	}

	return remainContent[0:endPos]
}

// StrLimit 限制字符串长度
func StrLimit(text string, length int) string {
	if length <= 0 {
		return ""
	}

	textRune := []rune(text)
	if len(textRune) > length {
		return string(textRune[:length]) + "..."
	}
	return text
}

// StrLen 获取字符串长度（Unicode字符数）
func StrLen(text string) int {
	return utf8.RuneCountInString(text)
}

// UTF8DecodeRune 字符串转Rune数组
func UTF8DecodeRune(text string) []string {
	res := make([]string, 0, len(text))
	for _, r := range []rune(text) {
		res = append(res, string(r))
	}
	return res
}

// Split2Tags 分割标签
func Split2Tags(text string) []string {
	if text == "" {
		return []string{}
	}

	tags := tagsRegex.Split(text, -1)
	newTags := make([]string, 0, len(tags))
	mapTags := make(map[string]bool, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if len(tag) > 0 && !mapTags[tag] {
			newTags = append(newTags, tag)
			mapTags[tag] = true
		}
	}
	return newTags
}

// Base64Encode base64编码
func Base64Encode(content string) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(content)), "=")
}

// Base64Decode base64解码
func Base64Decode(content string) string {
	missingPadding := len(content) % 4
	if missingPadding > 0 {
		content += strings.Repeat("=", 4-missingPadding)
	}

	v, err := base64.URLEncoding.DecodeString(content)
	if err != nil {
		return content
	}
	return string(v)
}

// CompleteBase64URLSafe 完整base64 URL安全编码
func CompleteBase64URLSafe(s string) string {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	padding := len(s) % 4
	if padding > 0 {
		s += strings.Repeat("=", 4-padding)
	}

	return s
}

// CleanSpecialSymbols 清理特殊符号
func CleanSpecialSymbols(s string) string {
	s = specialRegex.ReplaceAllString(s, "")
	s = chineseRegex.ReplaceAllString(s, "")
	return s
}

// CleanNewline 清理换行符
func CleanNewline(s string) string {
	return newlineRegex.ReplaceAllString(s, "")
}

// StringToHexKey 转换字符串为十六进制键
func StringToHexKey(input string) string {
	runes := []rune(input)
	return RunesToHexKey(runes)
}

// RunesToHexKey 转换Runes为十六进制键
func RunesToHexKey(runes []rune) string {
	hexParts := make([]string, len(runes))
	for i, r := range runes {
		hexParts[i] = fmt.Sprintf("%X", r)
	}
	return strings.Join(hexParts, "-")
}

// ClearSpace 清理空格
func ClearSpace(s string) string {
	if len(s) == 0 {
		return s
	}

	s = strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, s)

	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = htmlSpaceRegex.ReplaceAllString(s, "><")
	s = spaceRegex.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// ParseHumanNum 解析人类可读的数字格式
func ParseHumanNum(s string) int64 {
	s = strings.ToLower(strings.TrimSpace(s))

	units := map[string]float64{
		"k": 1000,
		"m": 1000000,
		"b": 1000000000,
		"t": 1000000000000,
	}

	var unit float64 = 1
	for suffix, value := range units {
		if strings.HasSuffix(s, suffix) {
			unit = value
			s = strings.TrimSuffix(s, suffix)
			break
		}
	}

	s = nonNumberRegex.ReplaceAllString(s, "")

	number, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return int64(number * unit)
}

// ==================== 新增的实用函数 ====================

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome 检查是否为回文
func IsPalindrome(s string) bool {
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

// CamelToSnake 驼峰转蛇形
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// SnakeToCamel 蛇形转驼峰
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) <= 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}

// TitleCase 转换为标题格式
func TitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// RemoveAccents 移除重音符号
func RemoveAccents(s string) string {
	// 简单的重音符号映射
	replacer := strings.NewReplacer(
		"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
		"è", "e", "é", "e", "ê", "e", "ë", "e",
		"ì", "i", "í", "i", "î", "i", "ï", "i",
		"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o",
		"ù", "u", "ú", "u", "û", "u", "ü", "u",
		"ý", "y", "ÿ", "y",
		"ç", "c", "ñ", "n",
		"À", "A", "Á", "A", "Â", "A", "Ã", "A", "Ä", "A", "Å", "A",
		"È", "E", "É", "E", "Ê", "E", "Ë", "E",
		"Ì", "I", "Í", "I", "Î", "I", "Ï", "I",
		"Ò", "O", "Ó", "O", "Ô", "O", "Õ", "O", "Ö", "O",
		"Ù", "U", "Ú", "U", "Û", "U", "Ü", "U",
		"Ý", "Y", "Ÿ", "Y",
		"Ç", "C", "Ñ", "N",
	)
	return replacer.Replace(s)
}

// CountWords 计算单词数量
func CountWords(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return len(strings.Fields(s))
}

// CountChars 计算字符数量（不包括空格）
func CountChars(s string) int {
	return len(strings.ReplaceAll(s, " ", ""))
}

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsPhone 验证电话号码格式
func IsPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// IsIPv4 验证IPv4地址格式
func IsIPv4(ip string) bool {
	return ipv4Regex.MatchString(ip)
}

// Truncate 截断字符串到指定长度
func Truncate(s string, length int, suffix ...string) string {
	if length <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= length {
		return s
	}

	suf := "..."
	if len(suffix) > 0 {
		suf = suffix[0]
	}

	return string(runes[:length]) + suf
}

// PadLeft 左填充字符串
func PadLeft(s string, length int, pad rune) string {
	runes := []rune(s)
	if len(runes) >= length {
		return s
	}

	padding := make([]rune, length-len(runes))
	for i := range padding {
		padding[i] = pad
	}

	return string(padding) + s
}

// PadRight 右填充字符串
func PadRight(s string, length int, pad rune) string {
	runes := []rune(s)
	if len(runes) >= length {
		return s
	}

	padding := make([]rune, length-len(runes))
	for i := range padding {
		padding[i] = pad
	}

	return s + string(padding)
}

// IsBlank 检查字符串是否为空或只包含空白字符
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotBlank 检查字符串是否不为空且不只包含空白字符
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// DefaultIfBlank 如果字符串为空则返回默认值
func DefaultIfBlank(s, defaultValue string) string {
	if IsBlank(s) {
		return defaultValue
	}
	return s
}

// IsASCII 检查字符串是否只包含ASCII字符
func IsASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// ContainsAny 检查字符串是否包含任意一个子字符串
func ContainsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll 检查字符串是否包含所有子字符串
func ContainsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// LevenshteinDistance 计算编辑距离
func LevenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,                             // deletion
				min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost), // insertion and substitution
			)
		}
	}

	return matrix[len1][len2]
}
