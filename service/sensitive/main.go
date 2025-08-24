package sensitive

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"net/http/httputil"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/mylukin/sensitive"
)

var (
	instances sync.Map
)

type SensitiveFilter struct {
	filter         atomic.Value // 存储 *sensitive.Filter
	dictURL        string
	updateInterval time.Duration
	client         *http.Client
	ctx            context.Context
	cancel         context.CancelFunc
	lastModified   string
	logMaxLines    int  // 日志输出的最大行数
	debug          bool // 新增：调试模式标志
}

type Option func(*SensitiveFilter)

func WithUpdateInterval(interval time.Duration) Option {
	return func(sf *SensitiveFilter) {
		sf.updateInterval = interval
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(sf *SensitiveFilter) {
		sf.client = client
	}
}

func WithLogMaxLines(maxLines int) Option {
	return func(sf *SensitiveFilter) {
		sf.logMaxLines = maxLines
	}
}

func WithDebugMode(debug bool) Option {
	return func(sf *SensitiveFilter) {
		sf.debug = debug
	}
}

func New(dictURL string, options ...Option) (*SensitiveFilter, error) {
	if dictURL == "" {
		return nil, errors.New("dictURL is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	sf := &SensitiveFilter{
		dictURL:        dictURL,
		updateInterval: 1 * time.Hour, // 默认更新间隔
		client:         &http.Client{Timeout: 10 * time.Second},
		ctx:            ctx,
		cancel:         cancel,
		logMaxLines:    50,    // 默认值设为 50
		debug:          false, // 默认关闭调试模式
	}

	for _, option := range options {
		option(sf)
	}

	if err := sf.updateDict(); err != nil {
		log.Warnf("Failed to initialize dictionary from URL, using empty filter: %v", err)
		// 不返回错误，继续使用空过滤器
		// 初始化一个空的过滤器作为默认值
		defaultFilter := sensitive.New()
		sf.filter.Store(defaultFilter)
	}

	go sf.autoUpdate()

	return sf, nil
}

func Get(dictURL string, options ...Option) (*SensitiveFilter, error) {
	// 尝试从 sync.Map 中获取实例
	if instance, ok := instances.Load(dictURL); ok {
		return instance.(*SensitiveFilter), nil
	}

	// 创建新实例
	instance, err := New(dictURL, options...)
	if err != nil {
		return nil, err
	}

	// 使用 LoadOrStore 确保并发安全
	actual, loaded := instances.LoadOrStore(dictURL, instance)
	if loaded {
		// 如果在创建新实例的过程中，另一个 goroutine 已经创建了实例，
		// 我们应该使用已存在的实例并关闭新创建的实例
		instance.Close()
		return actual.(*SensitiveFilter), nil
	}

	return instance, nil
}

func (sf *SensitiveFilter) checkUpdate() (bool, error) {
	req, err := http.NewRequestWithContext(sf.ctx, http.MethodHead, sf.dictURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HEAD request: %v", err)
	}

	if sf.lastModified != "" {
		req.Header.Set("If-Modified-Since", sf.lastModified)
	}

	// 确保请求方法是 HEAD
	req.Method = http.MethodHead

	// Log HTTP request before sending
	if sf.debug {
		// 仅在调试模式下执行
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Errorf("Failed to dump request: %v", err)
		} else {
			sf.logHTTP(reqDump, true)
		}
	}

	resp, err := sf.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send HEAD request: %v", err)
	}
	defer resp.Body.Close()

	// Log HTTP response (limited to 50 lines)
	if sf.debug {
		// 仅在调试模式下执行
		respDump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			log.Errorf("Failed to dump response: %v", err)
		} else {
			sf.logHTTP(respDump, false)
		}
	}

	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HEAD request failed, status code: %d", resp.StatusCode)
	}

	newLastModified := resp.Header.Get("Last-Modified")
	if newLastModified != "" && newLastModified != sf.lastModified {
		sf.lastModified = newLastModified
		return true, nil
	}

	return false, nil
}

func (sf *SensitiveFilter) updateDict() error {
	req, err := http.NewRequestWithContext(sf.ctx, http.MethodGet, sf.dictURL, nil)
	if err != nil {
		log.Errorf("Failed to create request: %v", err)
		return err
	}

	if sf.lastModified != "" {
		req.Header.Set("If-Modified-Since", sf.lastModified)
	}

	// Log HTTP request
	if sf.debug {
		// 仅在调试模式下执行
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Errorf("Failed to dump request: %v", err)
		} else {
			sf.logHTTP(reqDump, true)
		}
	}

	resp, err := sf.client.Do(req)
	if err != nil {
		log.Errorf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Log HTTP response (limited to 50 lines)
	if sf.debug {
		// 仅在调试模式下执行
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Errorf("Failed to dump response: %v", err)
		} else {
			sf.logHTTP(respDump, false)
		}
	}

	if resp.StatusCode == http.StatusNotModified {
		log.Info("Dictionary not modified, no update needed")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Failed to fetch dictionary, status code: %d", resp.StatusCode)
		return errors.New("failed to fetch dictionary")
	}

	newLastModified := resp.Header.Get("Last-Modified")
	if newLastModified != "" {
		sf.lastModified = newLastModified
	}

	newFilter := sensitive.New()
	if err := newFilter.Load(resp.Body); err != nil {
		log.Errorf("Failed to load new dictionary: %v", err)
		return err
	}

	sf.filter.Store(newFilter)

	log.Info("Dictionary updated successfully")
	return nil
}

func (sf *SensitiveFilter) autoUpdate() {
	ticker := time.NewTicker(sf.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			needUpdate, err := sf.checkUpdate()
			if err != nil {
				log.Errorf("Failed to check for updates: %v %s", err, sf.dictURL)
				continue
			}
			if needUpdate {
				if err := sf.updateDict(); err != nil {
					log.Errorf("Failed to update dictionary: %v", err)
				}
			}
		case <-sf.ctx.Done():
			return
		}
	}
}

func (sf *SensitiveFilter) Close() {
	sf.cancel()
}

func (sf *SensitiveFilter) getFilter() *sensitive.Filter {
	if sf == nil {
		log.Warn("SensitiveFilter is nil")
		return nil
	}
	filter := sf.filter.Load()
	if filter == nil {
		log.Warn("filter is nil, returning nil")
		return nil
	}
	return filter.(*sensitive.Filter)
}

func (sf *SensitiveFilter) AddWord(words ...string) {
	if len(words) == 0 {
		return
	}
	filter := sf.getFilter()
	if filter == nil {
		return
	}
	filter.AddWord(words...)
}

func (sf *SensitiveFilter) DelWord(words ...string) {
	if len(words) == 0 {
		return
	}
	filter := sf.getFilter()
	if filter == nil {
		return
	}
	filter.DelWord(words...)
}

func (sf *SensitiveFilter) Filter(text string) string {
	filter := sf.getFilter()
	if filter == nil {
		return text
	}
	return filter.Filter(text)
}

func (sf *SensitiveFilter) Replace(text string, repl rune) string {
	filter := sf.getFilter()
	if filter == nil {
		return text
	}
	return filter.Replace(text, repl)
}

func (sf *SensitiveFilter) FindIn(text string) (bool, string) {
	filter := sf.getFilter()
	if filter == nil {
		return false, ""
	}
	return filter.FindIn(text)
}

func (sf *SensitiveFilter) Validate(text string) (bool, string) {
	filter := sf.getFilter()
	if filter == nil {
		return true, ""
	}
	return filter.Validate(text)
}

func (sf *SensitiveFilter) FindAll(text string) []string {
	filter := sf.getFilter()
	if filter == nil {
		return []string{}
	}
	return filter.FindAll(text)
}

func (sf *SensitiveFilter) UpdateNoisePattern(pattern string) {
	currentFilter := sf.getFilter()
	if currentFilter == nil {
		return
	}
	newFilter := sensitive.New()
	*newFilter = *currentFilter // 复制所有字段
	newFilter.UpdateNoisePattern(pattern)
	sf.filter.Store(newFilter)
}

func (sf *SensitiveFilter) Length() int64 {
	filter := sf.getFilter()
	if filter == nil {
		return 0
	}
	return filter.Length()
}

func (sf *SensitiveFilter) logResponse(respDump []byte) {
	lines := strings.SplitN(string(respDump), "\n", sf.logMaxLines+1)
	if len(lines) > sf.logMaxLines {
		lines = lines[:sf.logMaxLines]
		lines = append(lines, fmt.Sprintf("... (truncated after %d lines)", sf.logMaxLines))
	}
	fmt.Printf("HTTP Response (first %d lines):\n%s\n", sf.logMaxLines, strings.Join(lines, "\n"))
}

func (sf *SensitiveFilter) logHTTP(data []byte, isRequest bool) {
	if !sf.debug {
		return
	}

	logType := "HTTP Response"
	if isRequest {
		logType = "HTTP Request"
	}

	lines := strings.SplitN(string(data), "\n", sf.logMaxLines+1)
	if len(lines) > sf.logMaxLines {
		lines = lines[:sf.logMaxLines]
		lines = append(lines, fmt.Sprintf("... (truncated after %d lines)", sf.logMaxLines))
	}
	fmt.Printf("%s (first %d lines):\n%s\n", logType, sf.logMaxLines, strings.Join(lines, "\n"))
}
