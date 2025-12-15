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
package api

import (
	"fmt"
	"net"
	"time"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// validateIPv4Format validates that a string is a valid IPv4 address.
func validateIPv4Format(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP address format")
	}
	// Ensure it's IPv4
	if parsed.To4() == nil {
		return fmt.Errorf("not a valid IPv4 address")
	}
	return nil
}

// validateIPv6Format validates that a string is a valid IPv6 address.
func validateIPv6Format(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP address format")
	}
	// Ensure it's IPv6 (To4() returns nil for IPv6)
	if parsed.To4() != nil {
		return fmt.Errorf("not a valid IPv6 address")
	}
	return nil
}

// validateDateTimeFormat validates that a string is a valid RFC 3339 date-time.
func validateDateTimeFormat(dt time.Time) error {
	// time.Time already validates during JSON unmarshaling
	// Check if it's zero value which might indicate parsing issues
	if dt.IsZero() {
		return fmt.Errorf("invalid date-time format")
	}
	return nil
}

// validateDevice validates format fields in a Device struct.
func validateDevice(device models.Device) error {
	// Validate IPv4 addresses
	if device.Ipv4Address != nil {
		if device.Ipv4Address.PublicAddress != nil {
			if err := validateIPv4Format(*device.Ipv4Address.PublicAddress); err != nil {
				return fmt.Errorf("invalid publicAddress: %w", err)
			}
		}
		if device.Ipv4Address.PrivateAddress != nil {
			if err := validateIPv4Format(*device.Ipv4Address.PrivateAddress); err != nil {
				return fmt.Errorf("invalid privateAddress: %w", err)
			}
		}
	}

	// Validate IPv6 address
	if device.Ipv6Address != nil {
		if err := validateIPv6Format(*device.Ipv6Address); err != nil {
			return fmt.Errorf("invalid ipv6Address: %w", err)
		}
	}

	return nil
}

// validateSinkCredential validates format fields in sink credentials.
func validateSinkCredential(cred *models.SinkCredential) error {
	if cred == nil {
		return nil
	}

	// Validate accessTokenExpiresUtc format
	if err := validateDateTimeFormat(cred.AccessTokenExpiresUtc); err != nil {
		return fmt.Errorf("invalid accessTokenExpiresUtc: %w", err)
	}

	return nil
}

// validatePowerSavingRequest validates all format fields in the PowerSavingRequest.
func validatePowerSavingRequest(req *models.PowerSavingRequest) error {
	// Validate devices
	for i, device := range req.Devices {
		if err := validateDevice(device); err != nil {
			return fmt.Errorf("device at index %d: %w", i, err)
		}
	}

	// Validate subscription request sink credential
	if req.SubscriptionRequest.SinkCredential != nil {
		if err := validateSinkCredential(req.SubscriptionRequest.SinkCredential); err != nil {
			return fmt.Errorf("subscriptionRequest.sinkCredential: %w", err)
		}
	}

	return nil
}
