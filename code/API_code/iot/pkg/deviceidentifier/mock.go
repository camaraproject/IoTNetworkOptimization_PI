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
package deviceidentifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

type MockTranslator struct{}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

func (m *MockTranslator) ResolveNetworkAccessIdentifier(ctx context.Context, device models.Device) (models.NetworkAccessIdentifier, error) {
	log := logger.Get()

	if device.NetworkAccessIdentifier != nil && *device.NetworkAccessIdentifier != "" {
		nai := *device.NetworkAccessIdentifier
		log.Debug("Using provided networkAccessIdentifier",
			zap.String("nai", string(nai)))
		return nai, nil
	}

	hashNAI, err := m.hashDeviceIdentifiers(device)
	if err != nil {
		log.Error("Failed to hash device identifiers", zap.Error(err))
		return "", fmt.Errorf("failed to hash device identifiers: %w", err)
	}

	log.Debug("Generated hash-based NAI from device identifiers",
		zap.String("nai", string(hashNAI)),
		zap.Bool("hasPhoneNumber", device.PhoneNumber != nil),
		zap.Bool("hasIpv4", device.Ipv4Address != nil),
		zap.Bool("hasIpv6", device.Ipv6Address != nil))

	return hashNAI, nil
}

// hashDeviceIdentifiers creates a deterministic hash-based NAI from device identifiers.
func (m *MockTranslator) hashDeviceIdentifiers(device models.Device) (models.NetworkAccessIdentifier, error) {
	identifierData := map[string]interface{}{}

	if device.PhoneNumber != nil && *device.PhoneNumber != "" {
		identifierData["phoneNumber"] = *device.PhoneNumber
	}
	if device.Ipv4Address != nil {
		identifierData["ipv4Address"] = device.Ipv4Address
	}
	if device.Ipv6Address != nil && *device.Ipv6Address != "" {
		identifierData["ipv6Address"] = *device.Ipv6Address
	}

	jsonData, err := json.Marshal(identifierData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal identifiers: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	hashStr := hex.EncodeToString(hash[:])

	nai := models.NetworkAccessIdentifier(fmt.Sprintf("%s@generated.nai", hashStr[:16]))

	return nai, nil
}
