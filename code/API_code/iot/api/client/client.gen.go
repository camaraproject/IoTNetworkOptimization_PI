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
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"

	. "github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ActivatePowerSavingWithBody request with any body
	ActivatePowerSavingWithBody(ctx context.Context, params *ActivatePowerSavingParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ActivatePowerSaving(ctx context.Context, params *ActivatePowerSavingParams, body ActivatePowerSavingJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetPowerSaving request
	GetPowerSaving(ctx context.Context, transactionId TransactionId, params *GetPowerSavingParams, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ActivatePowerSavingWithBody(ctx context.Context, params *ActivatePowerSavingParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewActivatePowerSavingRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ActivatePowerSaving(ctx context.Context, params *ActivatePowerSavingParams, body ActivatePowerSavingJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewActivatePowerSavingRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetPowerSaving(ctx context.Context, transactionId TransactionId, params *GetPowerSavingParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetPowerSavingRequest(c.Server, transactionId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewActivatePowerSavingRequest calls the generic ActivatePowerSaving builder with application/json body
func NewActivatePowerSavingRequest(server string, params *ActivatePowerSavingParams, body ActivatePowerSavingJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewActivatePowerSavingRequestWithBody(server, params, "application/json", bodyReader)
}

// NewActivatePowerSavingRequestWithBody generates requests for ActivatePowerSaving with any type of body
func NewActivatePowerSavingRequestWithBody(server string, params *ActivatePowerSavingParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/features/power-saving")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	if params != nil {

		if params.XCorrelator != nil {
			var headerParam0 string

			headerParam0, err = runtime.StyleParamWithLocation("simple", false, "x-correlator", runtime.ParamLocationHeader, *params.XCorrelator)
			if err != nil {
				return nil, err
			}

			req.Header.Set("x-correlator", headerParam0)
		}

	}

	return req, nil
}

// NewGetPowerSavingRequest generates requests for GetPowerSaving
func NewGetPowerSavingRequest(server string, transactionId TransactionId, params *GetPowerSavingParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "transactionId", runtime.ParamLocationPath, transactionId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/features/power-saving/transactions/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if params != nil {

		if params.XCorrelator != nil {
			var headerParam0 string

			headerParam0, err = runtime.StyleParamWithLocation("simple", false, "x-correlator", runtime.ParamLocationHeader, *params.XCorrelator)
			if err != nil {
				return nil, err
			}

			req.Header.Set("x-correlator", headerParam0)
		}

	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ActivatePowerSavingWithBodyWithResponse request with any body
	ActivatePowerSavingWithBodyWithResponse(ctx context.Context, params *ActivatePowerSavingParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ActivatePowerSavingResponse, error)

	ActivatePowerSavingWithResponse(ctx context.Context, params *ActivatePowerSavingParams, body ActivatePowerSavingJSONRequestBody, reqEditors ...RequestEditorFn) (*ActivatePowerSavingResponse, error)

	// GetPowerSavingWithResponse request
	GetPowerSavingWithResponse(ctx context.Context, transactionId TransactionId, params *GetPowerSavingParams, reqEditors ...RequestEditorFn) (*GetPowerSavingResponse, error)
}

type ActivatePowerSavingResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON202      *PowerSavingResponse
	JSON400      *Generic400
	JSON401      *Generic401
	JSON403      *Generic403
	JSON404      *Generic404
	JSON409      *Generic409
	JSON501      *Generic501
}

// Status returns HTTPResponse.Status
func (r ActivatePowerSavingResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ActivatePowerSavingResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetPowerSavingResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *PowerSavingResponse
	JSON400      *Generic400
	JSON401      *Generic401
	JSON403      *Generic403
	JSON404      *Generic404
	JSON501      *Generic501
}

// Status returns HTTPResponse.Status
func (r GetPowerSavingResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetPowerSavingResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ActivatePowerSavingWithBodyWithResponse request with arbitrary body returning *ActivatePowerSavingResponse
func (c *ClientWithResponses) ActivatePowerSavingWithBodyWithResponse(ctx context.Context, params *ActivatePowerSavingParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ActivatePowerSavingResponse, error) {
	rsp, err := c.ActivatePowerSavingWithBody(ctx, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseActivatePowerSavingResponse(rsp)
}

func (c *ClientWithResponses) ActivatePowerSavingWithResponse(ctx context.Context, params *ActivatePowerSavingParams, body ActivatePowerSavingJSONRequestBody, reqEditors ...RequestEditorFn) (*ActivatePowerSavingResponse, error) {
	rsp, err := c.ActivatePowerSaving(ctx, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseActivatePowerSavingResponse(rsp)
}

// GetPowerSavingWithResponse request returning *GetPowerSavingResponse
func (c *ClientWithResponses) GetPowerSavingWithResponse(ctx context.Context, transactionId TransactionId, params *GetPowerSavingParams, reqEditors ...RequestEditorFn) (*GetPowerSavingResponse, error) {
	rsp, err := c.GetPowerSaving(ctx, transactionId, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetPowerSavingResponse(rsp)
}

// ParseActivatePowerSavingResponse parses an HTTP response from a ActivatePowerSavingWithResponse call
func ParseActivatePowerSavingResponse(rsp *http.Response) (*ActivatePowerSavingResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ActivatePowerSavingResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 202:
		var dest PowerSavingResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON202 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Generic400
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Generic401
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Generic403
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Generic404
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 409:
		var dest Generic409
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON409 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 501:
		var dest Generic501
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON501 = &dest

	}

	return response, nil
}

// ParseGetPowerSavingResponse parses an HTTP response from a GetPowerSavingWithResponse call
func ParseGetPowerSavingResponse(rsp *http.Response) (*GetPowerSavingResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetPowerSavingResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest PowerSavingResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Generic400
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest Generic401
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 403:
		var dest Generic403
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON403 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest Generic404
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 501:
		var dest Generic501
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON501 = &dest

	}

	return response, nil
}
