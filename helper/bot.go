package helper

import (
	"regexp"
	"strconv"
	"strings"
)

// IsBotToken is test is bot token
func IsBotToken(token string) bool {
	return regexp.MustCompile(`([\d]{5,15}:[\w-]{35})`).MatchString(token)
}

// GetBotToken is get bot token for content
func GetBotToken(content string) (string, bool) {
	if botToken := regexp.MustCompile(`([\d]{5,15}:[\w-]{35})`).FindString(content); len(botToken) > 0 {
		return botToken, true
	}
	return "", false
}

// GetBotID is get bot id
func GetBotID(token string) int64 {
	botID := ""
	if pos := strings.Index(token, ":"); pos > -1 {
		botID = token[0:pos]
	}
	iBotID, _ := strconv.ParseInt(botID, 10, 64)
	return iBotID
}

// IsCommand is check text is command
func IsCommand(text string) bool {
	return len(text) > 1 && text[0] == '/' && strings.Count(text, "/") == 1
}

// ClearLink 清理URL链接
func ClearTGLink(link string) string {
	link = regexp.MustCompile(`[\s\r\n]+`).ReplaceAllString(link, "")
	link = regexp.MustCompile(`(?i)https?://[^/]+/`).ReplaceAllString(link, "")
	link = regexp.MustCompile(`(?i)@|t\.me/|telegram\.me/|telegram\.dog/|telesco\.pe/`).ReplaceAllString(link, "")
	if strings.Contains(link, "?") {
		link = link[0:strings.Index(link, "?")]
	}
	if strings.Contains(link, "/") {
		link = link[0:strings.Index(link, "/")]
	}
	link = regexp.MustCompile(`[^\w]+`).ReplaceAllString(link, "")
	return strings.TrimSpace(link)
}
