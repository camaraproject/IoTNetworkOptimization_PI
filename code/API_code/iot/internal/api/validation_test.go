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
	"testing"
	"time"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

func TestValidateIPv4Format(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "valid IPv4 with zeros",
			ip:      "10.0.0.0",
			wantErr: false,
		},
		{
			name:    "valid public IPv4",
			ip:      "84.125.93.10",
			wantErr: false,
		},
		{
			name:    "invalid IPv4 - too many octets",
			ip:      "192.168.1.1.1",
			wantErr: true,
		},
		{
			name:    "invalid IPv4 - out of range",
			ip:      "256.168.1.1",
			wantErr: true,
		},
		{
			name:    "invalid IPv4 - letters",
			ip:      "abc.def.ghi.jkl",
			wantErr: true,
		},
		{
			name:    "IPv6 should fail",
			ip:      "2001:db8::1",
			wantErr: true,
		},
		{
			name:    "empty string",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv4Format(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIPv4Format() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPv6Format(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "valid IPv6",
			ip:      "2001:db8:85a3:8d3:1319:8a2e:370:7344",
			wantErr: false,
		},
		{
			name:    "valid IPv6 compressed",
			ip:      "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "valid IPv6 loopback",
			ip:      "::1",
			wantErr: false,
		},
		{
			name:    "invalid IPv6 - malformed",
			ip:      "2001:db8:::",
			wantErr: true,
		},
		{
			name:    "IPv4 should fail",
			ip:      "192.168.1.1",
			wantErr: true,
		},
		{
			name:    "empty string",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv6Format(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIPv6Format() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateTimeFormat(t *testing.T) {
	tests := []struct {
		name    string
		dt      time.Time
		wantErr bool
	}{
		{
			name:    "valid time",
			dt:      time.Now(),
			wantErr: false,
		},
		{
			name:    "valid past time",
			dt:      time.Date(2023, 7, 3, 12, 27, 8, 312000000, time.UTC),
			wantErr: false,
		},
		{
			name:    "zero time should fail",
			dt:      time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateTimeFormat(tt.dt)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateTimeFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDevice(t *testing.T) {
	validIPv4 := "192.168.1.1"
	invalidIPv4 := "999.999.999.999"
	validIPv6 := "2001:db8::1"
	invalidIPv6 := "not-an-ipv6"

	tests := []struct {
		name    string
		device  models.Device
		wantErr bool
	}{
		{
			name: "valid device with IPv4",
			device: models.Device{
				Ipv4Address: &models.DeviceIpv4Addr{
					PublicAddress: &validIPv4,
				},
			},
			wantErr: false,
		},
		{
			name: "valid device with IPv6",
			device: models.Device{
				Ipv6Address: &validIPv6,
			},
			wantErr: false,
		},
		{
			name: "invalid device with bad IPv4 public",
			device: models.Device{
				Ipv4Address: &models.DeviceIpv4Addr{
					PublicAddress: &invalidIPv4,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid device with bad IPv4 private",
			device: models.Device{
				Ipv4Address: &models.DeviceIpv4Addr{
					PrivateAddress: &invalidIPv4,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid device with bad IPv6",
			device: models.Device{
				Ipv6Address: &invalidIPv6,
			},
			wantErr: true,
		},
		{
			name:    "valid device with no IP addresses",
			device:  models.Device{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDevice(tt.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSinkCredential(t *testing.T) {
	validTime := time.Date(2023, 7, 3, 12, 27, 8, 312000000, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name    string
		cred    *models.SinkCredential
		wantErr bool
	}{
		{
			name:    "nil credential",
			cred:    nil,
			wantErr: false,
		},
		{
			name: "valid credential with valid time",
			cred: &models.SinkCredential{
				CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
				AccessTokenCredential: models.AccessTokenCredential{
					AccessTokenExpiresUtc: validTime,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid credential with zero time",
			cred: &models.SinkCredential{
				CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
				AccessTokenCredential: models.AccessTokenCredential{
					AccessTokenExpiresUtc: zeroTime,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSinkCredential(tt.cred)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSinkCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePowerSavingRequest(t *testing.T) {
	validIPv4 := "192.168.1.1"
	invalidIPv4 := "999.999.999.999"
	validIPv6 := "2001:db8::1"
	invalidIPv6 := "not-an-ipv6"
	validTime := time.Date(2023, 7, 3, 12, 27, 8, 312000000, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name    string
		req     *models.PowerSavingRequest
		wantErr bool
	}{
		{
			name: "valid request with IPv4 device",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv4Address: &models.DeviceIpv4Addr{
							PublicAddress: &validIPv4,
						},
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: &models.SinkCredential{
						CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
						AccessTokenCredential: models.AccessTokenCredential{
							AccessTokenExpiresUtc: validTime,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with IPv6 device",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv6Address: &validIPv6,
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: &models.SinkCredential{
						CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
						AccessTokenCredential: models.AccessTokenCredential{
							AccessTokenExpiresUtc: validTime,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid request with bad IPv4",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv4Address: &models.DeviceIpv4Addr{
							PublicAddress: &invalidIPv4,
						},
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: &models.SinkCredential{
						CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
						AccessTokenCredential: models.AccessTokenCredential{
							AccessTokenExpiresUtc: validTime,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid request with bad IPv6",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv6Address: &invalidIPv6,
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: &models.SinkCredential{
						CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
						AccessTokenCredential: models.AccessTokenCredential{
							AccessTokenExpiresUtc: validTime,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid request with zero time",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv4Address: &models.DeviceIpv4Addr{
							PublicAddress: &validIPv4,
						},
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: &models.SinkCredential{
						CredentialType: models.SinkCredentialCredentialTypeACCESSTOKEN,
						AccessTokenCredential: models.AccessTokenCredential{
							AccessTokenExpiresUtc: zeroTime,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid request with no sink credential",
			req: &models.PowerSavingRequest{
				Devices: []models.Device{
					{
						Ipv4Address: &models.DeviceIpv4Addr{
							PublicAddress: &validIPv4,
						},
					},
				},
				SubscriptionRequest: models.SubscriptionRequest{
					SinkCredential: nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePowerSavingRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePowerSavingRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
