package helper

import (
	"regexp"
	"strings"
)

// EscapeHTML is escape HTML
func EscapeHTML(content string) string {
	content = strings.ReplaceAll(content, `"`, `&quot;`)
	content = strings.ReplaceAll(content, ">", "&gt;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	return content
}

// ClearHTML is clear html
func ClearHTML(s string) string {
	s = regexp.MustCompile(`\n|\r\n`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`<style[^>]*>.+?<\/style>`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`<script[^>]*>.+?<\/script>`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`<\/?[^>]+\/?>`).ReplaceAllString(s, " ")
	return s
}

// GetLinks is get links
func GetLinks(s string) []string {
	urls := []string{}
	matches := regexp.MustCompile(`(?i)<a\s+href="([^"]+)"[^>]*>`).FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		for _, v := range matches {
			urls = append(urls, v[1])
		}
	}
	return urls
}
