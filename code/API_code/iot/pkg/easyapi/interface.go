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

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// DeviceConfig holds the device performance profile configuration.
type DeviceConfig struct {
	PpMaximumLatency      string `json:"ppMaximumLatency"`
	PpMaximumResponseTime string `json:"ppMaximumResponseTime"`
}

// Client defines the interface for interacting with device actuation APIs.
type Client interface {
	// GetDeviceConfig retrieves the current performance profile configuration of a device.
	GetDeviceConfig(ctx context.Context, device models.Device) (*DeviceConfig, error)

	// SetDeviceConfig applies performance profile configuration to a device.
	SetDeviceConfig(ctx context.Context, device models.Device, config *DeviceConfig) error
}
