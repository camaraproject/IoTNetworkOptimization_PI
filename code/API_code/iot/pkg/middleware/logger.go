/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package middleware

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

// ZapLogger returns an Echo middleware that logs basic request information
// (method, URI, status, user-agent, latency, and errors) using Zap.
func ZapLogger() echo.MiddlewareFunc {
	log := logger.Get()
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogError:     true,
		LogUserAgent: true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			if c.Path() == "/healthz" {
				return nil
			}

			fields := []zap.Field{
				zap.String("method", values.Method),
				zap.String("uri", values.URI),
				zap.Int("status", values.Status),
				zap.String("userAgent", values.UserAgent),
				zap.Int64("latencyMicroseconds", values.Latency.Microseconds()),
			}
			if values.Error != nil {
				fields = append(fields, zap.Error(values.Error))
			}
			log.Debug("request", fields...)
			return nil
		},
	},
	)
}

// DebugBodyLogger returns a middleware that logs request and response bodies when debug mode is enabled.
func DebugBodyLogger() echo.MiddlewareFunc {
	if !logger.IsDebug() {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	log := logger.Get()
	return middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		if c.Path() == "/healthz" {
			return
		}
		var reqBodyMap, resBodyMap map[string]any
		if len(reqBody) > 0 {
			if err := json.Unmarshal(reqBody, &reqBodyMap); err != nil {
				log.Warn("Failed to unmarshal request body", zap.Error(err))
			}
		}
		if len(resBody) > 0 {
			if err := json.Unmarshal(resBody, &resBodyMap); err != nil {
				log.Warn("Failed to unmarshal response body", zap.Error(err))
			}
		}

		log.Debug("request/response body dump",
			zap.Any("reqBody", reqBodyMap),
			zap.Any("resBody", resBodyMap),
		)
	})
}
