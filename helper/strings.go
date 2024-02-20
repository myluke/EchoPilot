package helper

import (
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// FindSubstr 查找截取
func FindSubstr(content string, params ...interface{}) string {
	var result string = ""
	// 参数不全，直接返回
	if len(params) < 2 {
		return result
	}
	// 开始参数
	startArg := params[0]
	// 结束参数
	endArg := params[1]
	// 获取 start 出现的位置
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

	// 找不到开始位置，直接退出
	if startPos == -1 {
		return result
	}

	// 截取剩余内容
	remainContent := content[startPos:]

	// 获取 end 出现的位置
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

	// 找不到结束位置，直接退出
	if endPos == -1 {
		return result
	}

	// 找到了开始和结束位置，截取内容
	result = remainContent[0:endPos]
	return result
}

// StrLimit is string limit
func StrLimit(text string, length int) string {
	textRune := UTF8DecodeRune(text)
	textStrLen := len(text)
	textRealLen := len(textRune)
	// 英文，字数翻倍
	if textStrLen-textRealLen < textStrLen/2 {
		length *= 2
	}
	if textRealLen > length {
		text = strings.Join(textRune[0:length], "") + "..."
	}
	return text
}

// StrLen is get string length
func StrLen(text string) int {
	textLen := float64(utf8.RuneCountInString(text))
	byteLen := float64(len(text))
	// 英文，字数翻倍
	if byteLen-textLen < byteLen/2.8 {
		textLen = textLen / 2.8
	}
	return int(math.Floor(textLen))
}

// UTF8DecodeRune is string to Rune
func UTF8DecodeRune(text string) []string {
	var res []string
	bText := []byte(text)
	for len(bText) > 0 {
		r, size := utf8.DecodeRune(bText)
		if size == 4 {
			res = append(res, fmt.Sprintf("%c", r))
			res = append(res, "")
		} else {
			res = append(res, fmt.Sprintf("%c", r))
		}
		bText = bText[size:]
	}
	return res
}

// Split2Tags is split to tags
func Split2Tags(text string) []string {
	tags := regexp.MustCompile(`[,，、;；\|\s]\s*`).Split(text, -1)
	newTags := []string{}
	mapTags := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if len(tag) > 0 {
			// Deduplication
			if _, ok := mapTags[tag]; !ok {
				newTags = append(newTags, tag)
			}
			mapTags[tag] = true
		}
	}
	return newTags
}

// Base64Encode is base64 encode
func Base64Encode(content string) string {
	return base64.URLEncoding.EncodeToString([]byte(content))
}

// Base64Decode is base64 decode
func Base64Decode(content string) string {
	v, err := base64.URLEncoding.DecodeString(content)
	if err == nil {
		return string(v)
	}
	return content
}

// CompleteBase64URLSafe adjusts a Base64 URL-safe encoded string by replacing
// '-' with '+', '_' with '/', and adding missing padding '=' characters
func CompleteBase64URLSafe(s string) string {
	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Add padding if necessary
	padding := len(s) % 4
	if padding > 0 {
		s += strings.Repeat("=", 4-padding)
	}

	return s
}

// HiddenBotToken is hidden bot token
func HiddenBotToken(s string) string {
	return regexp.MustCompile(`\/(bot)?(\d+):([^/]+)\/?`).ReplaceAllString(s, "/$1$2:***********************************/")
}

// CleanSpecialSymbols is 清理特殊符号
func CleanSpecialSymbols(s string) string {
	s = regexp.MustCompile(`[\x00-\x1F]|[\x21-\x2F]|[\x3A-\x40]|[\x5B-\x60]|[\x7B-\x7F]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`[【】、，。？《》～！¥……（）——；‘：“「」｜]`).ReplaceAllString(s, "")
	return s
}

// CleanNewline is 清理换行符
func CleanNewline(s string) string {
	return regexp.MustCompile(`\r|\n|\t`).ReplaceAllString(s, "")
}

// StringToHexKey - Convert a string to hex string representation of their Unicode Code Point value
func StringToHexKey(input string) string {
	// Convert our input string to UTF runes
	runes := []rune(input)
	return RunesToHexKey(runes)
}

// RunesToHexKey - Convert a slice of runes to hex string representation of their Unicode Code Point value
func RunesToHexKey(runes []rune) string {
	// Build a slice of hex representations of each rune
	hexParts := []string{}
	for _, rune := range runes {
		hexParts = append(hexParts, fmt.Sprintf("%X", rune))
	}

	// Join the hex strings with a hypen - this is the key used in the emojis map
	return strings.Join(hexParts, "-")
}

// ClearSpace is clear space
func ClearSpace(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
}
