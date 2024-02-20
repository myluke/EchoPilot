package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	// PoweredByConfig is config
	PoweredByConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		Name string `yaml:"name"`

		Version string `yaml:"version"`
	}
)

// PoweredBy header add Powered-By
func PoweredBy(config PoweredByConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			res := c.Response()
			res.Header().Add("Powered-By", config.Name+"/"+config.Version)
			return next(c)
		}
	}
}
