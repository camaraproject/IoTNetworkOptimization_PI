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
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

var _ Client = &DummyClient{}

// DummyClient is a mock implementation of Client for testing.
type DummyClient struct{}

// NewDummy creates a new DummyClient.
func NewDummy() *DummyClient {
	return &DummyClient{}
}

// GetDeviceConfig returns simulated device performance profile configuration.
func (d *DummyClient) GetDeviceConfig(ctx context.Context, device models.Device) (*DeviceConfig, error) {
	log := logger.Get()

	deviceID := getDeviceIdentifier(device)
	log.Info("EASYAPI: Getting device configuration",
		zap.String("deviceId", deviceID),
		zap.Any("device", device))

	config := &DeviceConfig{
		PpMaximumLatency:      "100", // default non-power-saving value
		PpMaximumResponseTime: "200", // default non-power-saving value
	}

	log.Info("EASYAPI: Retrieved device configuration",
		zap.String("deviceId", deviceID),
		zap.Any("config", config))

	return config, nil
}

// SetDeviceConfig simulates applying performance profile configuration.
func (d *DummyClient) SetDeviceConfig(ctx context.Context, device models.Device, config *DeviceConfig) error {
	log := logger.Get()

	deviceID := getDeviceIdentifier(device)

	log.Info("EASYAPI: Setting device configuration",
		zap.String("deviceId", deviceID),
		zap.String("ppMaximumLatency", config.PpMaximumLatency),
		zap.String("ppMaximumResponseTime", config.PpMaximumResponseTime),
		zap.Any("device", device))

	log.Info("EASYAPI: Successfully applied device configuration",
		zap.String("deviceId", deviceID))

	return nil
}

// getDeviceIdentifier extracts a usable identifier from the Device object.
func getDeviceIdentifier(device models.Device) string {
	if device.PhoneNumber != nil {
		return *device.PhoneNumber
	}
	if device.Ipv4Address != nil {
		data, _ := json.Marshal(device.Ipv4Address)
		return fmt.Sprintf("ipv4:%s", string(data))
	}
	if device.Ipv6Address != nil {
		return fmt.Sprintf("ipv6:%s", *device.Ipv6Address)
	}
	if device.NetworkAccessIdentifier != nil {
		return *device.NetworkAccessIdentifier
	}
	return "unknown"
}
