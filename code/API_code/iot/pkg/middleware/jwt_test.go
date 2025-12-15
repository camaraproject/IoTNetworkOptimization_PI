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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// makeUnsignedJWT creates an unsigned JWT with the given claims.
func makeUnsignedJWT(claims map[string]interface{}) string {
	header := map[string]interface{}{
		"some": "header",
	}
	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(claims)
	headerEnc := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return headerEnc + "." + payloadEnc + "."
}

func TestMiddleware(t *testing.T) {
	type test struct {
		name        string
		claims      map[string]interface{}
		authHeader  string
		wantSub     string
		wantErrCode int
	}

	tests := []test{
		{
			name: "sets sub on context when present in the claim",
			claims: map[string]interface{}{
				"sub": "tom",
				"iat": 1234567890,
			},
			wantSub: "tom",
		},
		{
			name:        "fails when sub claim is missing",
			claims:      map[string]interface{}{},
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "fails when the auth header is invalid",
			authHeader:  "invalid",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "fails when JWT format is invalid",
			authHeader:  "Bearer invalid.token.parts",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "fails when base64 payload is invalid",
			authHeader:  "Bearer a.b.c",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authHeader := tt.authHeader
			if authHeader == "" && tt.claims != nil {
				tokenString := makeUnsignedJWT(tt.claims)
				authHeader = "Bearer " + tokenString
			}

			e := echo.New()
			var gotSub string
			h := func(c echo.Context) error {
				gotSub = CtxSub(c.Request().Context())
				return c.NoContent(http.StatusOK)
			}
			handler := JWT()(h)
			e.POST("/test", handler)

			req := httptest.NewRequest("POST", "/test", nil)
			req.Header.Set("Authorization", authHeader)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if tt.wantErrCode != 0 {
				assert.Equal(t, tt.wantErrCode, rec.Code)
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, tt.wantSub, gotSub)
			}
		})
	}
}
