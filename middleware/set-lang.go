package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mylukin/EchoPilot/helper"
	"github.com/mylukin/easy-i18n/i18n"
	"golang.org/x/text/language"
)

type (
	// SetLangConfig is config
	SetLangConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// 获取语言参数
		Language func(req *http.Request, res *echo.Response, c echo.Context) string
		// support languages
		Languages []language.Tag
	}
)

// SetLang set language
func SetLang(config SetLangConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			var tag language.Tag
			var accept string
			// 从函数里获取用户的语言设置
			if config.Language != nil {
				accept = config.Language(req, res, c)
			}
			// 获取不到语言设置，执行默认的语言设置
			if accept == "" {
				accept = helper.Config("LANGUAGE")
			}
			// 从URL里获取语言设置
			getLang := c.QueryParam("lang")
			if len(getLang) > 0 {
				tag = language.Make(getLang)
			} else {
				tag, _ = language.MatchStrings(language.NewMatcher(config.Languages), accept)
			}

			// 获取统一格式的语言设置
			userLang := tag.String()
			// 如果子语言，则取父级语言
			if strings.Count(userLang, "-") > 1 {
				userLang = tag.Parent().String()
			}
			// 全部变为小写
			userLang = strings.ToLower(userLang)
			// 设置环境变量
			c.Set("Language", i18n.NewPrinter(userLang))

			res.Header().Add("Language", userLang)

			return next(c)
		}
	}
}
