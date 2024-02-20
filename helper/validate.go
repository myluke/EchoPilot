package helper

import (
	"regexp"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsURL is test is url
func IsURL(token string) bool {
	return regexp.MustCompile(`^(?i)https?://[\w\-]+(\.[\w\-]+){1,}`).MatchString(token)
}

// IsObjectID is object id
func IsObjectID(v string) bool {
	return primitive.IsValidObjectID(v)
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
