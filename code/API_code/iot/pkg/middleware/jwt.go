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
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

type CtxKey string

const (
	Sub CtxKey = "sub"
)

func extractSubFromJWT(req *http.Request) (string, int, string) {
	log := logger.Get()
	reqToken := req.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		msg := "invalid Bearer token in Authorization header"
		log.Error(msg)
		return "", http.StatusBadRequest, msg
	}
	reqToken = splitToken[1]

	parts := strings.Split(reqToken, ".")
	if len(parts) < 2 {
		msg := "invalid JWT token format"
		log.Error(msg)
		return "", http.StatusBadRequest, msg
	}

	payloadSegment := parts[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		msg := "failed to decode JWT payload"
		log.With(zap.Error(err)).Error(msg)
		return "", http.StatusBadRequest, msg
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		msg := "failed to unmarshal JWT payload"
		log.With(zap.Error(err)).Error(msg)
		return "", http.StatusBadRequest, msg
	}

	if sub, ok := claims["sub"].(string); ok {
		return sub, 0, ""
	}

	return "", http.StatusBadRequest, "sub claim not found in JWT"
}

func JWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/healthz" {
				return next(c)
			}
			sub, status, msg := extractSubFromJWT(c.Request())
			if status != 0 {
				return c.String(status, msg)
			}
			ctx := context.WithValue(c.Request().Context(), Sub, sub)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func CtxSub(ctx context.Context) string {
	var sub string
	if s, ok := ctx.Value(Sub).(string); ok {
		sub = s
	}
	return sub
}
