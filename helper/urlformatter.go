package helper

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

// FormatURL 格式化HTML中的URL地址
// base: 页面地址
// html: html代码
// 返回处理后的html代码
func FormatURL(base string, html string) string {
	// 基础检查
	if base == "" || html == "" {
		return html
	}

	// 解析base URL
	baseURL, err := url.Parse(base)
	if err != nil {
		return html
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return html
	}

	// 构建基础URL信息
	baseHost := baseURL.Scheme + "://" + baseURL.Host
	basePath := baseURL.Path
	if hashIndex := strings.Index(basePath, "#"); hashIndex != -1 {
		basePath = basePath[:hashIndex]
	}
	basePath = path.Dir(basePath)
	if basePath == "/" {
		basePath = ""
	}
	baseFullURL := baseHost + basePath

	// 编译正则表达式（避免重复编译）
	re := regexp.MustCompile(`(?i)(href|src)=['"]?([^'">\s]+)['"]?`)

	// 存储所有需要替换的URL及其位置
	type urlMatch struct {
		start, end int
		oldURL     string
		newURL     string
	}
	var matches []urlMatch

	// 查找所有匹配
	for _, match := range re.FindAllStringSubmatchIndex(html, -1) {
		urlStart := match[4]
		urlEnd := match[5]
		urlStr := html[urlStart:urlEnd]

		// 跳过已经是完整URL或特殊URL的情况
		if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") ||
			strings.HasPrefix(urlStr, "ftp://") || strings.HasPrefix(urlStr, "data:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "mailto:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "javascript:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "tel:") ||
			strings.HasPrefix(urlStr, "#") {
			continue
		}

		// 处理URL
		var newURL string
		if strings.HasPrefix(urlStr, "//") {
			newURL = baseURL.Scheme + ":" + urlStr
		} else if strings.HasPrefix(urlStr, "/") {
			newURL = baseHost + urlStr
		} else if strings.HasPrefix(urlStr, "../") {
			tempPath := basePath
			tempURL := urlStr
			for strings.HasPrefix(tempURL, "../") {
				tempURL = tempURL[3:]
				if tempPath != "" {
					tempPath = path.Dir(tempPath)
					if tempPath == "/" {
						tempPath = ""
					}
				}
			}
			newURL = baseHost + tempPath + "/" + tempURL
		} else if strings.HasPrefix(urlStr, "./") {
			newURL = baseFullURL + "/" + urlStr[2:]
		} else {
			newURL = baseFullURL + "/" + urlStr
		}

		// 清理新URL
		newURL = strings.ReplaceAll(newURL, "//", "/")
		newURL = strings.Replace(newURL, ":/", "://", 1)
		newURL = strings.TrimSuffix(newURL, "/")

		// 保存匹配信息
		matches = append(matches, urlMatch{
			start:  urlStart,
			end:    urlEnd,
			oldURL: urlStr,
			newURL: newURL,
		})
	}

	// 从后向前替换，避免位置偏移
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		html = html[:m.start] + m.newURL + html[m.end:]
	}

	return html
}
