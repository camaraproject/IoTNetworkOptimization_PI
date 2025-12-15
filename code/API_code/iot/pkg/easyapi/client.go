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
package easyapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

var _ Client = &EasyApiClient{}

// EasyApiClient implements the Client interface using the EasyAPI backend.
type EasyApiClient struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new EasyAPI client.
func New(baseURL string) *EasyApiClient {
	return &EasyApiClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AccessAndMobilitySubscriptionData represents the 3GPP TS 29.503 AM data response.
type AccessAndMobilitySubscriptionData struct {
	SubsRegTimer *int `json:"subsRegTimer,omitempty"`
	ActiveTime   *int `json:"activeTime,omitempty"`
}

// GetDeviceConfig retrieves device configuration via GET /nudm-sdm/v2/{supi}/am-data.
func (c *EasyApiClient) GetDeviceConfig(ctx context.Context, device models.Device) (*DeviceConfig, error) {
	log := logger.Get()

	if device.NetworkAccessIdentifier == nil {
		return nil, fmt.Errorf("networkAccessIdentifier is required")
	}

	supi := string(*device.NetworkAccessIdentifier)
	url := fmt.Sprintf("%s/nudm-sdm/v2/%s/am-data", c.baseURL, url.PathEscape(supi))

	log.Info("EasyAPI: Getting device AM data",
		zap.String("url", url),
		zap.String("supi", supi))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Error("Failed to create HTTP request",
			zap.String("supi", supi),
			zap.Error(err))
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to execute HTTP request",
			zap.String("supi", supi),
			zap.Error(err))
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body",
			zap.String("supi", supi),
			zap.Error(err))
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("EasyAPI returned non-200 status",
			zap.String("supi", supi),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("EasyAPI error: status %d", resp.StatusCode)
	}

	var amData AccessAndMobilitySubscriptionData
	if err := json.Unmarshal(body, &amData); err != nil {
		log.Error("Failed to parse AM data response",
			zap.String("supi", supi),
			zap.Error(err))
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Info("EasyAPI: Retrieved AM data",
		zap.String("supi", supi),
		zap.Any("amData", amData))

	config := &DeviceConfig{}

	if amData.SubsRegTimer != nil {
		config.PpMaximumLatency = fmt.Sprintf("%d", *amData.SubsRegTimer)
	} else {
		log.Error("Missing required field in AM data response",
			zap.String("supi", supi),
			zap.String("field", "subsRegTimer"))
		return nil, fmt.Errorf("subsRegTimer field is missing in response")
	}

	if amData.ActiveTime != nil {
		config.PpMaximumResponseTime = fmt.Sprintf("%d", *amData.ActiveTime)
	} else {
		log.Error("Missing required field in AM data response",
			zap.String("supi", supi),
			zap.String("field", "activeTime"))
		return nil, fmt.Errorf("activeTime field is missing in response")
	}

	log.Info("EasyAPI: Mapped to device config",
		zap.String("supi", supi),
		zap.Any("config", config))

	return config, nil
}

// PpDataUpdate represents the request body for updating PP data.
type PpDataUpdate struct {
	PpData *PpDataPayload `json:"ppData"`
}

// PpDataPayload contains the PP data configuration.
type PpDataPayload struct {
	CommunicationCharacteristics *CommunicationCharacteristicsUpdate `json:"communicationCharacteristics"`
}

// CommunicationCharacteristicsUpdate contains the communication characteristics to update.
type CommunicationCharacteristicsUpdate struct {
	PpMaximumLatency      *string `json:"ppMaximumLatency,omitempty"`
	PpMaximumResponseTime *string `json:"ppMaximumResponseTime,omitempty"`
}

// SetDeviceConfig updates device configuration via PATCH /nudm-pp/v1/{ueId}/pp-data.
func (c *EasyApiClient) SetDeviceConfig(ctx context.Context, device models.Device, config *DeviceConfig) error {
	log := logger.Get()

	if device.NetworkAccessIdentifier == nil {
		return fmt.Errorf("networkAccessIdentifier is required")
	}

	ueId := string(*device.NetworkAccessIdentifier)
	url := fmt.Sprintf("%s/nudm-pp/v1/%s/pp-data", c.baseURL, url.PathEscape(ueId))

	log.Info("EasyAPI: Setting device configuration",
		zap.String("url", url),
		zap.String("ueId", ueId),
		zap.Any("config", config))

	requestBody := PpDataUpdate{
		PpData: &PpDataPayload{
			CommunicationCharacteristics: &CommunicationCharacteristicsUpdate{
				PpMaximumLatency:      &config.PpMaximumLatency,
				PpMaximumResponseTime: &config.PpMaximumResponseTime,
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Error("Failed to marshal request body",
			zap.String("ueId", ueId),
			zap.Error(err))
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Error("Failed to create HTTP request",
			zap.String("ueId", ueId),
			zap.Error(err))
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to execute HTTP request",
			zap.String("ueId", ueId),
			zap.Error(err))
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body",
			zap.String("ueId", ueId),
			zap.Error(err))
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		log.Error("EasyAPI returned non-success status",
			zap.String("ueId", ueId),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("body", string(body)))
		return fmt.Errorf("EasyAPI error: status %d", resp.StatusCode)
	}

	log.Info("EasyAPI: Device configuration updated successfully",
		zap.String("ueId", ueId))

	return nil
}
