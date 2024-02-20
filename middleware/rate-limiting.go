package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// RateLimitingConfig is config
	RateLimitingConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// 生成ID
		Generator func(req *http.Request, res *echo.Response, c echo.Context) string
		// 窗口时间，单位：毫秒
		Window time.Duration `yaml:"window"`
		// 最大请求次数
		Limit int `yaml:"limit"`
		// Cache
		Cache map[string]*RateLimitingCache
		// 回调函数
		Callback func(req *http.Request, res *echo.Response, c echo.Context) error
		// lock
		sync.RWMutex
	}
	// RateLimitingCache is cache
	RateLimitingCache struct {
		Value   int
		Expired time.Time
	}
)

// RateLimiting is rate limit
func RateLimiting(config *RateLimitingConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	if len(config.Cache) == 0 {
		config.Cache = map[string]*RateLimitingCache{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			rqeustID := config.Generator(req, res, c)
			if rqeustID == "" {
				return next(c)
			}

			// 加锁
			config.Lock()
			if cache, ok := config.Cache[rqeustID]; ok {
				// cache no expired
				if cache.Expired.UnixNano() >= time.Now().UnixNano() {
					newValue := cache.Value + 1
					config.Cache[rqeustID].Value = newValue
					if newValue > config.Limit {
						// Increase expiration time
						config.Cache[rqeustID].Expired = time.Now().Add(config.Window * time.Millisecond)
						// 解锁
						config.Unlock()
						// 返回 429
						return config.Callback(req, res, c)
					}
				}
			} else {
				config.Cache[rqeustID] = &RateLimitingCache{
					Value:   1,
					Expired: time.Now().Add(config.Window * time.Millisecond),
				}
			}
			config.Unlock()

			res.Header().Set(echo.HeaderXRequestID, rqeustID)

			return next(c)
		}
	}
}
