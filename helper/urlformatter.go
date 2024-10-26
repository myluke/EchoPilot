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
// 返回处理后的html代码和可能的错误
func FormatURL(base string, html string) string {
	// 参数验证
	if base == "" || html == "" {
		return html
	}

	// 验证并规范化base URL
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
	// 如果baseURL指向文件，获取其目录
	if !strings.HasSuffix(basePath, "/") && strings.Contains(path.Base(basePath), ".") {
		basePath = path.Dir(basePath)
	}
	if basePath == "/" {
		basePath = ""
	}
	baseFullURL := baseHost + basePath

	// 使用两个单独的正则表达式匹配
	srcRe := regexp.MustCompile(`<(img|script)[^>]+src=['"]?([^'">\s]+)['"]?[^>]*>`)
	hrefRe := regexp.MustCompile(`<(a|link)[^>]+href=['"]?([^'">\s]+)['"]?[^>]*>`)

	// 处理所有匹配
	processRegex := func(re *regexp.Regexp, html string) string {
		matches := re.FindAllStringSubmatchIndex(html, -1)
		if matches == nil {
			return html
		}

		// 从后向前处理匹配项，避免替换时的位置偏移问题
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			urlStart := match[4]
			urlEnd := match[5]

			// 提取和清理URL
			urlStr := strings.TrimSpace(html[urlStart:urlEnd])
			if urlStr == "" {
				continue
			}

			// 跳过特殊URL
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
				// 处理协议相对URL
				newURL = baseHost[:strings.Index(baseHost, ":")+1] + urlStr
			} else if strings.HasPrefix(urlStr, "/") {
				// 处理绝对路径
				newURL = baseHost + urlStr
			} else if strings.HasPrefix(urlStr, "../") {
				// 处理相对路径 ../
				tempBasePath := basePath
				tempURL := urlStr
				for strings.HasPrefix(tempURL, "../") {
					tempURL = tempURL[3:]
					if tempBasePath != "" {
						tempBasePath = path.Dir(tempBasePath)
						if tempBasePath == "/" {
							tempBasePath = ""
						}
					}
					if tempURL == "../" {
						tempURL = ""
						break
					}
				}
				newURL = baseHost + tempBasePath + "/" + tempURL
			} else if strings.HasPrefix(urlStr, "./") {
				// 处理当前路径 ./
				newURL = baseFullURL + "/" + urlStr[2:]
			} else {
				// 处理其他情况
				newURL = baseFullURL + "/" + urlStr
			}

			// 清理URL
			newURL = regexp.MustCompile(`/+`).ReplaceAllString(newURL, "/")
			if idx := strings.Index(newURL, "://"); idx != -1 {
				protocol := newURL[:idx+3]
				rest := newURL[idx+3:]
				newURL = protocol + strings.TrimLeft(rest, "/")
			}
			if !strings.HasSuffix(newURL, "://") && strings.HasSuffix(newURL, "/") {
				newURL = strings.TrimRight(newURL, "/")
			}

			// 替换原始URL，保持原有的引号
			prefix := html[match[0]:urlStart]
			suffix := html[urlEnd:match[1]]
			html = html[:match[0]] + prefix + newURL + suffix + html[match[1]:]
		}
		return html
	}

	// 依次处理src和href属性
	html = processRegex(srcRe, html)
	html = processRegex(hrefRe, html)

	return html
}
