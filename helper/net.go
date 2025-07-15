package helper

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 预编译正则表达式以提高性能
var (
	htmlTagRegex      = regexp.MustCompile(`<\/?[^>]+\/?>`)
	styleRegex        = regexp.MustCompile(`<style[^>]*>.+?<\/style>`)
	scriptRegex       = regexp.MustCompile(`<script[^>]*>.+?<\/script>`)
	linkRegex         = regexp.MustCompile(`(?i)<a\s+href="([^"]+)"[^>]*>`)
	urlRegex          = regexp.MustCompile(`^(?i)https?://[\w\-]+(\.[\w\-]+){1,}`)
	emailRegex2       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	ipv4Regex2        = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	ipv6Regex         = regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
	macRegex          = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	domainRegex       = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	portRegex         = regexp.MustCompile(`^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`)
	botTokenRegex     = regexp.MustCompile(`\/(bot)?(\d+):([^/]+)\/?`)
	tgLinkRegex       = regexp.MustCompile(`[\s\r\n]+`)
	tgLinkPrefixRegex = regexp.MustCompile(`(?i)https?://[^/]+/`)
	tgLinkHandleRegex = regexp.MustCompile(`(?i)@|t\.me/|telegram\.me/|telegram\.dog/|telesco\.pe/`)
	tgLinkCleanRegex  = regexp.MustCompile(`[^\w]+`)
)

// ==================== HTML 处理 ====================

// EscapeHTML HTML转义
func EscapeHTML(content string) string {
	content = strings.ReplaceAll(content, `"`, `&quot;`)
	content = strings.ReplaceAll(content, ">", "&gt;")
	content = strings.ReplaceAll(content, "<", "&lt;")
	return content
}

// UnescapeHTML HTML反转义
func UnescapeHTML(content string) string {
	content = strings.ReplaceAll(content, `&quot;`, `"`)
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&amp;", "&")
	return content
}

// ClearHTML 清理HTML标签
func ClearHTML(s string) string {
	s = regexp.MustCompile(`\n|\r\n`).ReplaceAllString(s, "")
	s = styleRegex.ReplaceAllString(s, "")
	s = scriptRegex.ReplaceAllString(s, "")
	s = htmlTagRegex.ReplaceAllString(s, " ")
	return s
}

// StripHTML 移除HTML标签
func StripHTML(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}

// GetLinks 从HTML中提取链接
func GetLinks(s string) []string {
	urls := []string{}
	matches := linkRegex.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		for _, v := range matches {
			urls = append(urls, v[1])
		}
	}
	return urls
}

// ExtractText 从HTML中提取纯文本
func ExtractText(html string) string {
	// 移除script和style标签
	html = styleRegex.ReplaceAllString(html, "")
	html = scriptRegex.ReplaceAllString(html, "")

	// 移除HTML标签
	html = htmlTagRegex.ReplaceAllString(html, " ")

	// 清理空白字符
	html = regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")

	return strings.TrimSpace(html)
}

// ==================== URL 处理 ====================

// FormatURL 格式化HTML中的URL地址
func FormatURL(base string, html string) string {
	if base == "" || html == "" {
		return html
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return html
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return html
	}

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

	re := regexp.MustCompile(`(?i)(href|src)=['"]?([^'">\s]+)['"]?`)

	type urlMatch struct {
		start, end int
		oldURL     string
		newURL     string
	}
	var matches []urlMatch

	for _, match := range re.FindAllStringSubmatchIndex(html, -1) {
		urlStart := match[4]
		urlEnd := match[5]
		urlStr := html[urlStart:urlEnd]

		if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") ||
			strings.HasPrefix(urlStr, "ftp://") || strings.HasPrefix(urlStr, "data:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "mailto:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "javascript:") ||
			strings.HasPrefix(strings.ToLower(urlStr), "tel:") ||
			strings.HasPrefix(urlStr, "#") {
			continue
		}

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

		newURL = strings.ReplaceAll(newURL, "//", "/")
		newURL = strings.Replace(newURL, ":/", "://", 1)
		newURL = strings.TrimSuffix(newURL, "/")

		matches = append(matches, urlMatch{
			start:  urlStart,
			end:    urlEnd,
			oldURL: urlStr,
			newURL: newURL,
		})
	}

	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		html = html[:m.start] + m.newURL + html[m.end:]
	}

	return html
}

// ParseURL 解析URL
func ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// BuildURL 构建URL
func BuildURL(scheme, host, path string, params map[string]string) string {
	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// JoinURL 连接URL路径
func JoinURL(base, rel string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	relURL, err := url.Parse(rel)
	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(relURL).String(), nil
}

// GetURLParams 获取URL参数
func GetURLParams(rawURL string) (map[string]string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for k, v := range u.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	return params, nil
}

// AddURLParam 添加URL参数
func AddURLParam(rawURL, key, value string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// RemoveURLParam 移除URL参数
func RemoveURLParam(rawURL, key string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// GetDomain 获取域名
func GetDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return u.Hostname(), nil
}

// GetPort 获取端口
func GetPort(rawURL string) (int, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, err
	}

	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "http":
			return 80, nil
		case "https":
			return 443, nil
		case "ftp":
			return 21, nil
		default:
			return 0, fmt.Errorf("无法确定端口")
		}
	}

	return strconv.Atoi(port)
}

// ==================== 验证函数 ====================

// IsURL 验证URL格式
func IsURL(token string) bool {
	return urlRegex.MatchString(token)
}

// IsValidURL 验证URL是否有效
func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsHTTPURL 验证是否为HTTP/HTTPS URL
func IsHTTPURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// IsSecureURL 验证是否为HTTPS URL
func IsSecureURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// IsNetEmail 验证邮箱格式
func IsNetEmail(email string) bool {
	return emailRegex2.MatchString(email)
}

// IsNetIPv4 验证IPv4地址
func IsNetIPv4(ip string) bool {
	return ipv4Regex2.MatchString(ip)
}

// IsIPv6 验证IPv6地址
func IsIPv6(ip string) bool {
	return ipv6Regex.MatchString(ip)
}

// IsIP 验证IP地址（IPv4或IPv6）
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsPrivateIP 验证是否为私有IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 私有地址范围
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// IsPublicIP 验证是否为公网IP
func IsPublicIP(ip string) bool {
	return IsIP(ip) && !IsPrivateIP(ip)
}

// IsMAC 验证MAC地址
func IsMAC(mac string) bool {
	return macRegex.MatchString(mac)
}

// IsDomain 验证域名
func IsDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

// IsPort 验证端口号
func IsPort(port string) bool {
	return portRegex.MatchString(port)
}

// IsLocalhost 验证是否为本地地址
func IsLocalhost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// ==================== IP 工具 ====================

// ParseIP 解析IP地址
func ParseIP(ip string) net.IP {
	return net.ParseIP(ip)
}

// ParseCIDR 解析CIDR
func ParseCIDR(cidr string) (net.IP, *net.IPNet, error) {
	return net.ParseCIDR(cidr)
}

// IsIPInRange 检查IP是否在CIDR范围内
func IsIPInRange(ip, cidr string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return network.Contains(parsedIP)
}

// GetIPType 获取IP类型
func GetIPType(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "invalid"
	}

	if parsedIP.To4() != nil {
		return "ipv4"
	}

	return "ipv6"
}

// GetNetworkIP 获取网络地址
func GetNetworkIP(ip, mask string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("无效的IP地址")
	}

	parsedMask := net.ParseIP(mask)
	if parsedMask == nil {
		return "", fmt.Errorf("无效的子网掩码")
	}

	ipv4 := parsedIP.To4()
	maskv4 := parsedMask.To4()

	if ipv4 == nil || maskv4 == nil {
		return "", fmt.Errorf("仅支持IPv4")
	}

	networkIP := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		networkIP[i] = ipv4[i] & maskv4[i]
	}

	return networkIP.String(), nil
}

// GetBroadcastIP 获取广播地址
func GetBroadcastIP(ip, mask string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("无效的IP地址")
	}

	parsedMask := net.ParseIP(mask)
	if parsedMask == nil {
		return "", fmt.Errorf("无效的子网掩码")
	}

	ipv4 := parsedIP.To4()
	maskv4 := parsedMask.To4()

	if ipv4 == nil || maskv4 == nil {
		return "", fmt.Errorf("仅支持IPv4")
	}

	broadcastIP := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcastIP[i] = ipv4[i] | (^maskv4[i])
	}

	return broadcastIP.String(), nil
}

// ==================== Bot 相关 ====================

// HiddenBotToken 隐藏bot token
func HiddenBotToken(s string) string {
	return botTokenRegex.ReplaceAllString(s, "/$1$2:***********************************/")
}

// IsBotToken 检测是否是bot token
func IsBotToken(token string) bool {
	return regexp.MustCompile(`([\d]{5,15}:[\w-]{35})`).MatchString(token)
}

// GetBotToken 从内容中获取bot token
func GetBotToken(content string) (string, bool) {
	if botToken := regexp.MustCompile(`([\d]{5,15}:[\w-]{35})`).FindString(content); len(botToken) > 0 {
		return botToken, true
	}
	return "", false
}

// GetBotID 从token中获取bot ID
func GetBotID(token string) int64 {
	botID := ""
	if pos := strings.Index(token, ":"); pos > -1 {
		botID = token[0:pos]
	}
	iBotID, _ := strconv.ParseInt(botID, 10, 64)
	return iBotID
}

// IsCommand 检查文本是否是命令
func IsCommand(text string) bool {
	return len(text) > 1 && text[0] == '/' && strings.Count(text, "/") == 1
}

// ClearTGLink 清理TG链接
func ClearTGLink(link string) string {
	link = tgLinkRegex.ReplaceAllString(link, "")
	link = tgLinkPrefixRegex.ReplaceAllString(link, "")
	link = tgLinkHandleRegex.ReplaceAllString(link, "")
	if strings.Contains(link, "?") {
		link = link[0:strings.Index(link, "?")]
	}
	if strings.Contains(link, "/") {
		link = link[0:strings.Index(link, "/")]
	}
	link = tgLinkCleanRegex.ReplaceAllString(link, "")
	return strings.TrimSpace(link)
}

// ==================== 工具函数 ====================

// URLEncode URL编码
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// URLDecode URL解码
func URLDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// NormalizeURL 标准化URL
func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 转换为小写
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// 移除默认端口
	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = u.Hostname()
	}

	// 移除尾部斜杠
	if u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}

// SanitizeURL 清理URL（移除危险字符）
func SanitizeURL(rawURL string) string {
	// 移除危险字符
	dangerous := []string{
		"<", ">", "\"", "'", "&", "\n", "\r", "\t",
	}

	result := rawURL
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "")
	}

	return result
}

// IsAbsoluteURL 检查是否为绝对URL
func IsAbsoluteURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// IsRelativeURL 检查是否为相对URL
func IsRelativeURL(rawURL string) bool {
	return !IsAbsoluteURL(rawURL)
}

// GetURLExtension 获取URL文件扩展名
func GetURLExtension(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	ext := path.Ext(u.Path)
	if ext != "" {
		return ext[1:] // 移除点号
	}

	return ""
}

// GetURLFilename 获取URL文件名
func GetURLFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return path.Base(u.Path)
}

// GetURLPath 获取URL路径
func GetURLPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Path
}

// GetURLQuery 获取URL查询字符串
func GetURLQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.RawQuery
}

// GetURLFragment 获取URL片段
func GetURLFragment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Fragment
}

// ==================== 网络连接工具 ====================

// PingHost 检查主机是否可达
func PingHost(host string, timeout int) bool {
	conn, err := net.DialTimeout("tcp", host, time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetLocalIPs 获取本地IP地址
func GetLocalIPs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// GetOutboundIP 获取出站IP
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int, timeout int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetFreePort 获取空闲端口
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
