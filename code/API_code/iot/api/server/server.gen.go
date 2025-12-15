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
package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"

	. "github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Activate or de-activate power-saving features for IoT devices
	// (POST /features/power-saving)
	ActivatePowerSaving(ctx echo.Context, params ActivatePowerSavingParams) error
	// Get status of a transaction for power-saving features
	// (GET /features/power-saving/transactions/{transactionId})
	GetPowerSaving(ctx echo.Context, transactionId TransactionId, params GetPowerSavingParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ActivatePowerSaving converts echo context to params.
func (w *ServerInterfaceWrapper) ActivatePowerSaving(ctx echo.Context) error {
	var err error

	ctx.Set(OAuth2Scopes, []string{"iot-management:power-saving:write"})

	// Parameter object where we will unmarshal all parameters from the context
	var params ActivatePowerSavingParams

	headers := ctx.Request().Header
	// ------------- Optional header parameter "x-correlator" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("x-correlator")]; found {
		var XCorrelator XCorrelator
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for x-correlator, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "x-correlator", valueList[0], &XCorrelator, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: false})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter x-correlator: %s", err))
		}

		params.XCorrelator = &XCorrelator
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.ActivatePowerSaving(ctx, params)
	return err
}

// GetPowerSaving converts echo context to params.
func (w *ServerInterfaceWrapper) GetPowerSaving(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "transactionId" -------------
	var transactionId TransactionId

	err = runtime.BindStyledParameterWithOptions("simple", "transactionId", ctx.Param("transactionId"), &transactionId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter transactionId: %s", err))
	}

	ctx.Set(OAuth2Scopes, []string{"iot-management:power-saving:read"})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetPowerSavingParams

	headers := ctx.Request().Header
	// ------------- Optional header parameter "x-correlator" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("x-correlator")]; found {
		var XCorrelator XCorrelator
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for x-correlator, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "x-correlator", valueList[0], &XCorrelator, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: false})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter x-correlator: %s", err))
		}

		params.XCorrelator = &XCorrelator
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetPowerSaving(ctx, transactionId, params)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/features/power-saving", wrapper.ActivatePowerSaving)
	router.GET(baseURL+"/features/power-saving/transactions/:transactionId", wrapper.GetPowerSaving)

}

type Generic400ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic400JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic400ResponseHeaders
}

type Generic401ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic401JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic401ResponseHeaders
}

type Generic403ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic403JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic403ResponseHeaders
}

type Generic404ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic404JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic404ResponseHeaders
}

type Generic409ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic409JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic409ResponseHeaders
}

type Generic410ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic410JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic410ResponseHeaders
}

type Generic429ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic429JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic429ResponseHeaders
}

type Generic501ResponseHeaders struct {
	XCorrelator XCorrelator
}
type Generic501JSONResponse struct {
	Body struct {
		Code interface{} `json:"code"`

		// Message Detailed error description
		Message string      `json:"message"`
		Status  interface{} `json:"status"`
	}

	Headers Generic501ResponseHeaders
}

type ActivatePowerSavingRequestObject struct {
	Params ActivatePowerSavingParams
	Body   *ActivatePowerSavingJSONRequestBody
}

type ActivatePowerSavingResponseObject interface {
	VisitActivatePowerSavingResponse(w http.ResponseWriter) error
}

type ActivatePowerSaving202ResponseHeaders struct {
	XCorrelator XCorrelator
}

type ActivatePowerSaving202JSONResponse struct {
	Body    PowerSavingResponse
	Headers ActivatePowerSaving202ResponseHeaders
}

func (response ActivatePowerSaving202JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(202)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving400JSONResponse struct{ Generic400JSONResponse }

func (response ActivatePowerSaving400JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving401JSONResponse struct{ Generic401JSONResponse }

func (response ActivatePowerSaving401JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving403JSONResponse struct{ Generic403JSONResponse }

func (response ActivatePowerSaving403JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving404JSONResponse struct{ Generic404JSONResponse }

func (response ActivatePowerSaving404JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving409JSONResponse struct{ Generic409JSONResponse }

func (response ActivatePowerSaving409JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(409)

	return json.NewEncoder(w).Encode(response.Body)
}

type ActivatePowerSaving501JSONResponse struct{ Generic501JSONResponse }

func (response ActivatePowerSaving501JSONResponse) VisitActivatePowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(501)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetPowerSavingRequestObject struct {
	TransactionId TransactionId `json:"transactionId"`
	Params        GetPowerSavingParams
}

type GetPowerSavingResponseObject interface {
	VisitGetPowerSavingResponse(w http.ResponseWriter) error
}

type GetPowerSaving200JSONResponse PowerSavingResponse

func (response GetPowerSaving200JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetPowerSaving400JSONResponse struct{ Generic400JSONResponse }

func (response GetPowerSaving400JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetPowerSaving401JSONResponse struct{ Generic401JSONResponse }

func (response GetPowerSaving401JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(401)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetPowerSaving403JSONResponse struct{ Generic403JSONResponse }

func (response GetPowerSaving403JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetPowerSaving404JSONResponse struct{ Generic404JSONResponse }

func (response GetPowerSaving404JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(404)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetPowerSaving501JSONResponse struct{ Generic501JSONResponse }

func (response GetPowerSaving501JSONResponse) VisitGetPowerSavingResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-correlator", fmt.Sprint(response.Headers.XCorrelator))
	w.WriteHeader(501)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Activate or de-activate power-saving features for IoT devices
	// (POST /features/power-saving)
	ActivatePowerSaving(ctx context.Context, request ActivatePowerSavingRequestObject) (ActivatePowerSavingResponseObject, error)
	// Get status of a transaction for power-saving features
	// (GET /features/power-saving/transactions/{transactionId})
	GetPowerSaving(ctx context.Context, request GetPowerSavingRequestObject) (GetPowerSavingResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// ActivatePowerSaving operation middleware
func (sh *strictHandler) ActivatePowerSaving(ctx echo.Context, params ActivatePowerSavingParams) error {
	var request ActivatePowerSavingRequestObject

	request.Params = params

	var body ActivatePowerSavingJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.ActivatePowerSaving(ctx.Request().Context(), request.(ActivatePowerSavingRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "ActivatePowerSaving")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(ActivatePowerSavingResponseObject); ok {
		return validResponse.VisitActivatePowerSavingResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetPowerSaving operation middleware
func (sh *strictHandler) GetPowerSaving(ctx echo.Context, transactionId TransactionId, params GetPowerSavingParams) error {
	var request GetPowerSavingRequestObject

	request.TransactionId = transactionId
	request.Params = params

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetPowerSaving(ctx.Request().Context(), request.(GetPowerSavingRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetPowerSaving")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetPowerSavingResponseObject); ok {
		return validResponse.VisitGetPowerSavingResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"GmsAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
