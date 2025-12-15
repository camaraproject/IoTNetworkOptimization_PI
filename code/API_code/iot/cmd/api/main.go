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
package main

import (
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
	"go.uber.org/zap"

	handler "github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/api"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/middleware"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/server"
)

func main() {
	conf := config.GetConf()
	log := logger.Get()

	e := echo.New()

	// Liveness/readiness endpoint for Knative/K8s probes (registered first, before any middleware)
	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.Use(middleware.DebugBodyLogger())
	e.Use(middleware.ZapLogger())
	e.Use(middleware.JWT())

	// Load OpenAPI spec from file for validation
	specPath := os.Getenv("OPENAPI_SPEC_PATH")
	if specPath == "" {
		specPath = "./api/iot-network-optimization.yaml"
	}

	// Load OpenAPI spec for validation
	swagger, err := server.GetSwagger()
	if err != nil {
		log.With(zap.Error(err)).Fatal("failed to load OpenAPI spec")
	}
	// Skip server validation - we don't want to validate server URLs
	swagger.Servers = nil

	// Add OpenAPI validation middleware for request validation (skip healthz)
	e.Use(echomiddleware.OapiRequestValidatorWithOptions(swagger, &echomiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
		Skipper: func(c echo.Context) bool {
			// Skip validation for health check endpoint
			return c.Path() == "/healthz"
		},
	}))

	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		log.With(zap.Error(err), zap.String("Mongo URI", conf.Database.Uri), zap.String("DB Name", conf.Database.Name)).
			Fatal("failed to connect to mongo Database")
	}

	h, err := handler.New(db)
	if err != nil {
		log.With(zap.Error(err)).
			Fatal("failed to create api handler")
	}
	server.RegisterHandlers(e, h)

	log.Info("Starting server", zap.String("address", conf.API.Address))
	if err := e.Start(conf.API.Address); err != nil {
		log.With(zap.Error(err)).
			Fatal("failed to run server")
	}
}
