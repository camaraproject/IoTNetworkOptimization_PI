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
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	OAuth2Scopes = "oAuth2.Scopes"
)

// Defines values for AccessTokenCredentialAccessTokenType.
const (
	AccessTokenCredentialAccessTokenTypeBearer AccessTokenCredentialAccessTokenType = "bearer"
)

// Defines values for AccessTokenCredentialCredentialType.
const (
	AccessTokenCredentialCredentialTypeACCESSTOKEN  AccessTokenCredentialCredentialType = "ACCESSTOKEN"
	AccessTokenCredentialCredentialTypePLAIN        AccessTokenCredentialCredentialType = "PLAIN"
	AccessTokenCredentialCredentialTypeREFRESHTOKEN AccessTokenCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for CloudEventDatacontenttype.
const (
	Applicationjson CloudEventDatacontenttype = "application/json"
)

// Defines values for CloudEventSpecversion.
const (
	N10 CloudEventSpecversion = "1.0"
)

// Defines values for DeviceStatusStatus.
const (
	Failed     DeviceStatusStatus = "failed"
	InProgress DeviceStatusStatus = "in-progress"
	Pending    DeviceStatusStatus = "pending"
	Success    DeviceStatusStatus = "success"
)

// Defines values for EventTypeNotification.
const (
	EventTypeNotificationOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving      EventTypeNotification = "org.camaraproject.iot-network-optimization-notification.v1.power-saving"
	EventTypeNotificationOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError EventTypeNotification = "org.camaraproject.iot-network-optimization-notification.v1.power-saving.error"
)

// Defines values for HTTPSettingsMethod.
const (
	POST HTTPSettingsMethod = "POST"
)

// Defines values for PlainCredentialCredentialType.
const (
	PlainCredentialCredentialTypeACCESSTOKEN  PlainCredentialCredentialType = "ACCESSTOKEN"
	PlainCredentialCredentialTypePLAIN        PlainCredentialCredentialType = "PLAIN"
	PlainCredentialCredentialTypeREFRESHTOKEN PlainCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for Protocol.
const (
	AMQP  Protocol = "AMQP"
	HTTP  Protocol = "HTTP"
	KAFKA Protocol = "KAFKA"
	MQTT3 Protocol = "MQTT3"
	MQTT5 Protocol = "MQTT5"
	NATS  Protocol = "NATS"
)

// Defines values for RefreshTokenCredentialAccessTokenType.
const (
	RefreshTokenCredentialAccessTokenTypeBearer RefreshTokenCredentialAccessTokenType = "bearer"
)

// Defines values for RefreshTokenCredentialCredentialType.
const (
	ACCESSTOKEN  RefreshTokenCredentialCredentialType = "ACCESSTOKEN"
	PLAIN        RefreshTokenCredentialCredentialType = "PLAIN"
	REFRESHTOKEN RefreshTokenCredentialCredentialType = "REFRESHTOKEN"
)

// Defines values for SubscriptionEventType.
const (
	SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving      SubscriptionEventType = "org.camaraproject.iot-network-optimization-notification.v1.power-saving"
	SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError SubscriptionEventType = "org.camaraproject.iot-network-optimization-notification.v1.power-saving.error"
)

// AccessTokenCredential defines model for AccessTokenCredential.
type AccessTokenCredential struct {
	// AccessToken REQUIRED. An access token is a previously acquired token granting access to the target resource.
	AccessToken string `json:"accessToken"`

	// AccessTokenExpiresUtc REQUIRED. An absolute (UTC) timestamp at which the token shall be considered expired.
	// In the case of an ACCESS_TOKEN_EXPIRED termination reason, implementation should notify the client before the expiration date.
	// If the access token is a JWT and registered "exp" (Expiration Time) claim is present, the two expiry times should match.
	// It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	AccessTokenExpiresUtc time.Time `json:"accessTokenExpiresUtc"`

	// AccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
	AccessTokenType AccessTokenCredentialAccessTokenType `json:"accessTokenType"`

	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType AccessTokenCredentialCredentialType `json:"credentialType"`
}

// AccessTokenCredentialAccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
type AccessTokenCredentialAccessTokenType string

// AccessTokenCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type AccessTokenCredentialCredentialType string

// CloudEvent The notification callback
type CloudEvent struct {
	// Data Event details payload described in each CAMARA API and referenced by its type
	Data *map[string]interface{} `json:"data,omitempty"`

	// Datacontenttype media-type that describes the event payload encoding, must be "application/json" for CAMARA APIs
	Datacontenttype *CloudEventDatacontenttype `json:"datacontenttype,omitempty"`

	// Id identifier of this event, that must be unique in the source context.
	Id string `json:"id"`

	// Source Identifies the context in which an event happened - be a non-empty
	// `URI-reference` like:
	// - URI with a DNS authority:
	//   * https://github.com/cloudevents
	//   * mailto:cncf-wg-serverless@lists.cncf.io
	// - Universally-unique URN with a UUID:
	//   * urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66
	// - Application-specific identifier:
	//   * /cloudevents/spec/pull/123
	//   * 1-555-123-4567
	Source Source `json:"source"`

	// Specversion Version of the specification to which this event conforms (must be 1.0 if it conforms to cloudevents 1.0.2 version)
	Specversion CloudEventSpecversion `json:"specversion"`

	// Time Timestamp of when the occurrence happened. Must adhere to RFC 3339.
	// WARN: This optional field in CloudEvents specification is required in
	// CAMARA APIs implementation.
	Time DateTime `json:"time"`

	// Type Event triggered when an event-type event occurred.
	Type EventTypeNotification `json:"type"`
}

// CloudEventDatacontenttype media-type that describes the event payload encoding, must be "application/json" for CAMARA APIs
type CloudEventDatacontenttype string

// CloudEventSpecversion Version of the specification to which this event conforms (must be 1.0 if it conforms to cloudevents 1.0.2 version)
type CloudEventSpecversion string

// CloudEventError The notification callback
type CloudEventError = CloudEvent

// CloudEventPowerSaving The notification callback
type CloudEventPowerSaving = CloudEvent

// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
// Specific event type attributes must be defined in `subscriptionDetail`
// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
type Config struct {
	// InitialEvent Set to `true` by API consumer if consumer wants to get an event as soon as the subscription is created and current situation reflects event request.
	// Example: Consumer request Roaming event. If consumer sets initialEvent to true and device is in roaming situation, an event is triggered
	// Up to API project decision to keep it.
	InitialEvent *bool `json:"initialEvent,omitempty"`

	// SubscriptionDetail The detail of the requested event subscription.
	SubscriptionDetail CreateSubscriptionDetail `json:"subscriptionDetail"`

	// SubscriptionExpireTime The subscription expiration time (in date-time format) requested by the API consumer. It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone. Up to API project decision to keep it.
	SubscriptionExpireTime *time.Time `json:"subscriptionExpireTime,omitempty"`

	// SubscriptionMaxEvents Identifies the maximum number of event reports to be generated (>=1) requested by the API consumer - Once this number is reached, the subscription ends. Up to API project decision to keep it.
	SubscriptionMaxEvents *int `json:"subscriptionMaxEvents,omitempty"`
}

// CreateSubscriptionDetail The detail of the requested event subscription.
type CreateSubscriptionDetail = map[string]interface{}

// DateTime Timestamp of when the occurrence happened. Must adhere to RFC 3339.
// WARN: This optional field in CloudEvents specification is required in
// CAMARA APIs implementation.
type DateTime = time.Time

// Device End-user equipment able to connect to a mobile network. Examples of
// devices include smartphones or IoT sensors/actuators.
// The developer can choose to provide the below specified device
// identifiers:
// * `ipv4Address`
// * `ipv6Address`
// * `phoneNumber`
// * `networkAccessIdentifier`
// NOTE1: the MNO might support only a subset of these options.
// The API invoker can provide multiple identifiers to be compatible
// across different MNOs. In this case the identifiers MUST belong to
// the same device.
// NOTE2: as for this Commonalities release, we are enforcing that the
// networkAccessIdentifier is only part of the schema for
// future-proofing, and CAMARA does not currently allow its use.
// After the CAMARA meta-release work is concluded and the relevant
// issues are resolved, its use will need to be explicitly documented
// in the guidelines.
type Device struct {
	// Ipv4Address The device should be identified by either the public (observed) IP
	//   address and port as seen by the application server, or the private
	//   (local) and any public (observed) IP addresses in use by the device
	//   (this information can be obtained by various means, for example from
	//   some DNS servers).
	// If the allocated and observed IP addresses are the same (i.e. NAT is not
	//   in use) then  the same address should be specified for both
	//   publicAddress and privateAddress.
	// If NAT64 is in use, the device should be identified by its publicAddress
	//   and publicPort, or separately by its allocated IPv6 address (field
	//   ipv6Address of the Device object)
	// In all cases, publicAddress must be specified, along with at least one
	//   of either privateAddress or publicPort, dependent upon which is known.
	//   In general, mobile devices cannot be identified by their public IPv4
	//   address alone.
	Ipv4Address *DeviceIpv4Addr `json:"ipv4Address,omitempty"`

	// Ipv6Address The device should be identified by the observed IPv6 address, or by any
	//   single IPv6 address from within the subnet allocated to the device
	//   (e.g. adding ::0 to the /64 prefix).
	Ipv6Address *DeviceIpv6Address `json:"ipv6Address,omitempty"`

	// NetworkAccessIdentifier A public identifier addressing a subscription in a mobile network. In 3GPP terminology, it corresponds to the GPSI formatted with the External Identifier ({Local Identifier}@{Domain Identifier}). Unlike the telephone number, the network access identifier is not subjected to portability ruling in force, and is individually managed by each operator.
	NetworkAccessIdentifier *NetworkAccessIdentifier `json:"networkAccessIdentifier,omitempty"`

	// PhoneNumber A public identifier addressing a telephone subscription. In mobile networks it corresponds to the MSISDN (Mobile Station International Subscriber Directory Number). In order to be globally unique it has to be formatted in international format, according to E.164 standard, prefixed with '+'.
	PhoneNumber *PhoneNumber `json:"phoneNumber,omitempty"`
}

// DeviceIpv4Addr The device should be identified by either the public (observed) IP
//
//	address and port as seen by the application server, or the private
//	(local) and any public (observed) IP addresses in use by the device
//	(this information can be obtained by various means, for example from
//	some DNS servers).
//
// If the allocated and observed IP addresses are the same (i.e. NAT is not
//
//	in use) then  the same address should be specified for both
//	publicAddress and privateAddress.
//
// If NAT64 is in use, the device should be identified by its publicAddress
//
//	and publicPort, or separately by its allocated IPv6 address (field
//	ipv6Address of the Device object)
//
// In all cases, publicAddress must be specified, along with at least one
//
//	of either privateAddress or publicPort, dependent upon which is known.
//	In general, mobile devices cannot be identified by their public IPv4
//	address alone.
type DeviceIpv4Addr struct {
	// PrivateAddress A single IPv4 address with no subnet mask
	PrivateAddress *SingleIpv4Addr `json:"privateAddress,omitempty"`

	// PublicAddress A single IPv4 address with no subnet mask
	PublicAddress *SingleIpv4Addr `json:"publicAddress,omitempty"`

	// PublicPort TCP or UDP port number
	PublicPort *Port `json:"publicPort,omitempty"`
	union      json.RawMessage
}

// DeviceIpv4Addr0 defines model for .
type DeviceIpv4Addr0 = interface{}

// DeviceIpv4Addr1 defines model for .
type DeviceIpv4Addr1 = interface{}

// DeviceIpv6Address The device should be identified by the observed IPv6 address, or by any
//
//	single IPv6 address from within the subnet allocated to the device
//	(e.g. adding ::0 to the /64 prefix).
type DeviceIpv6Address = string

// DeviceStatus defines model for DeviceStatus.
type DeviceStatus struct {
	// Device End-user equipment able to connect to a mobile network. Examples of
	// devices include smartphones or IoT sensors/actuators.
	// The developer can choose to provide the below specified device
	// identifiers:
	// * `ipv4Address`
	// * `ipv6Address`
	// * `phoneNumber`
	// * `networkAccessIdentifier`
	// NOTE1: the MNO might support only a subset of these options.
	// The API invoker can provide multiple identifiers to be compatible
	// across different MNOs. In this case the identifiers MUST belong to
	// the same device.
	// NOTE2: as for this Commonalities release, we are enforcing that the
	// networkAccessIdentifier is only part of the schema for
	// future-proofing, and CAMARA does not currently allow its use.
	// After the CAMARA meta-release work is concluded and the relevant
	// issues are resolved, its use will need to be explicitly documented
	// in the guidelines.
	Device *Device             `json:"device,omitempty"`
	Status *DeviceStatusStatus `json:"status,omitempty"`
}

// DeviceStatusStatus defines model for DeviceStatus.Status.
type DeviceStatusStatus string

// ErrorInfo error information
type ErrorInfo struct {
	// Code Code given to this error
	Code string `json:"code"`

	// Message Detailed error description
	Message string `json:"message"`

	// Status HTTP status code returned along with this error response
	Status int `json:"status"`
}

// EventTypeNotification Event triggered when an event-type event occurred.
type EventTypeNotification string

// HTTPSettings defines model for HTTPSettings.
type HTTPSettings struct {
	// Headers A set of key/value pairs that is copied into the HTTP request as custom headers.
	//
	// NOTE: Use/Applicability of this concept has not been discussed in Commonalities under the scope of Meta Release v0.4. When required by an API project as an option to meet a UC/Requirement, please generate an issue for Commonalities discussion about it.
	Headers *map[string]string `json:"headers,omitempty"`

	// Method The HTTP method to use for sending the message.
	Method *HTTPSettingsMethod `json:"method,omitempty"`
}

// HTTPSettingsMethod The HTTP method to use for sending the message.
type HTTPSettingsMethod string

// NetworkAccessIdentifier A public identifier addressing a subscription in a mobile network. In 3GPP terminology, it corresponds to the GPSI formatted with the External Identifier ({Local Identifier}@{Domain Identifier}). Unlike the telephone number, the network access identifier is not subjected to portability ruling in force, and is individually managed by each operator.
type NetworkAccessIdentifier = string

// PhoneNumber A public identifier addressing a telephone subscription. In mobile networks it corresponds to the MSISDN (Mobile Station International Subscriber Directory Number). In order to be globally unique it has to be formatted in international format, according to E.164 standard, prefixed with '+'.
type PhoneNumber = string

// PlainCredential defines model for PlainCredential.
type PlainCredential struct {
	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType PlainCredentialCredentialType `json:"credentialType"`

	// Identifier The identifier might be an account or username.
	Identifier string `json:"identifier"`

	// Secret The secret might be a password or passphrase.
	Secret string `json:"secret"`
}

// PlainCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type PlainCredentialCredentialType string

// Port TCP or UDP port number
type Port = int

// PowerSavingRequest defines model for PowerSavingRequest.
type PowerSavingRequest struct {
	// Devices Device IDs or group identifiers
	Devices []Device `json:"devices"`
	Enabled bool     `json:"enabled"`

	// SubscriptionRequest The request for creating a event-type event subscription
	SubscriptionRequest SubscriptionRequest `json:"subscriptionRequest"`
	TimePeriod          *struct {
		// EndDate An instant of time, ending of the TimePeriod. If not included, then the period has no ending date.
		EndDate *time.Time `json:"endDate,omitempty"`

		// StartDate An instant of time, starting of the TimePeriod.
		StartDate time.Time `json:"startDate"`
	} `json:"timePeriod,omitempty"`
}

// PowerSavingResponse defines model for PowerSavingResponse.
type PowerSavingResponse struct {
	ActivationStatus *[]DeviceStatus `json:"activationStatus,omitempty"`
	TransactionId    *string         `json:"transactionId,omitempty"`
}

// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
type Protocol string

// RefreshTokenCredential defines model for RefreshTokenCredential.
type RefreshTokenCredential struct {
	// AccessToken REQUIRED. An access token is a previously acquired token granting access to the target resource.
	AccessToken *string `json:"accessToken,omitempty"`

	// AccessTokenExpiresUtc REQUIRED. An absolute (UTC) timestamp at which the token shall be considered expired.
	// In the case of an ACCESS_TOKEN_EXPIRED termination reason, implementation should notify the client before the expiration date.
	// If the access token is a JWT and registered "exp" (Expiration Time) claim is present, the two expiry times should match.
	// It must follow [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339#section-5.6) and must have time zone.
	AccessTokenExpiresUtc *time.Time `json:"accessTokenExpiresUtc,omitempty"`

	// AccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
	AccessTokenType *RefreshTokenCredentialAccessTokenType `json:"accessTokenType,omitempty"`

	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType RefreshTokenCredentialCredentialType `json:"credentialType"`

	// RefreshToken REQUIRED. An refresh token credential used to acquire access tokens.
	RefreshToken *string `json:"refreshToken,omitempty"`

	// RefreshTokenEndpoint REQUIRED. A URL at which the refresh token can be traded for an access token.
	RefreshTokenEndpoint *string `json:"refreshTokenEndpoint,omitempty"`
}

// RefreshTokenCredentialAccessTokenType REQUIRED. Type of the access token (See [OAuth 2.0](https://tools.ietf.org/html/rfc6749#section-7.1)).
type RefreshTokenCredentialAccessTokenType string

// RefreshTokenCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type RefreshTokenCredentialCredentialType string

// SingleIpv4Addr A single IPv4 address with no subnet mask
type SingleIpv4Addr = string

// Source Identifies the context in which an event happened - be a non-empty
// `URI-reference` like:
// - URI with a DNS authority:
//   - https://github.com/cloudevents
//   - mailto:cncf-wg-serverless@lists.cncf.io
//
// - Universally-unique URN with a UUID:
//   - urn:uuid:6e8bc430-9c3a-11d9-9669-0800200c9a66
//
// - Application-specific identifier:
//   - /cloudevents/spec/pull/123
//   - 1-555-123-4567
type Source = string

// SubscriptionEventType event-type that could be subscribed through this subscription. Several event-type could be defined.
type SubscriptionEventType string

// TransactionId Transaction identifier allocated for enabling/disabling IoT features
type TransactionId = openapi_types.UUID

// XCorrelator defines model for XCorrelator.
type XCorrelator = string

// Generic400 defines model for Generic400.
type Generic400 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic401 defines model for Generic401.
type Generic401 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic403 defines model for Generic403.
type Generic403 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic404 defines model for Generic404.
type Generic404 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic409 defines model for Generic409.
type Generic409 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic410 defines model for Generic410.
type Generic410 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic422 defines model for Generic422.
type Generic422 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic429 defines model for Generic429.
type Generic429 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// Generic501 defines model for Generic501.
type Generic501 struct {
	Code interface{} `json:"code"`

	// Message Detailed error description
	Message string      `json:"message"`
	Status  interface{} `json:"status"`
}

// ActivatePowerSavingParams defines parameters for ActivatePowerSaving.
type ActivatePowerSavingParams struct {
	// XCorrelator Correlation id for the different services
	XCorrelator *XCorrelator `json:"x-correlator,omitempty"`
}

// GetPowerSavingParams defines parameters for GetPowerSaving.
type GetPowerSavingParams struct {
	// XCorrelator Correlation id for the different services
	XCorrelator *XCorrelator `json:"x-correlator,omitempty"`
}

// ActivatePowerSavingJSONRequestBody defines body for ActivatePowerSaving for application/json ContentType.
type ActivatePowerSavingJSONRequestBody = PowerSavingRequest

// AsDeviceIpv4Addr0 returns the union data inside the DeviceIpv4Addr as a DeviceIpv4Addr0
func (t DeviceIpv4Addr) AsDeviceIpv4Addr0() (DeviceIpv4Addr0, error) {
	var body DeviceIpv4Addr0
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDeviceIpv4Addr0 overwrites any union data inside the DeviceIpv4Addr as the provided DeviceIpv4Addr0
func (t *DeviceIpv4Addr) FromDeviceIpv4Addr0(v DeviceIpv4Addr0) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDeviceIpv4Addr0 performs a merge with any union data inside the DeviceIpv4Addr, using the provided DeviceIpv4Addr0
func (t *DeviceIpv4Addr) MergeDeviceIpv4Addr0(v DeviceIpv4Addr0) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsDeviceIpv4Addr1 returns the union data inside the DeviceIpv4Addr as a DeviceIpv4Addr1
func (t DeviceIpv4Addr) AsDeviceIpv4Addr1() (DeviceIpv4Addr1, error) {
	var body DeviceIpv4Addr1
	err := json.Unmarshal(t.union, &body)
	return body, err
}

// FromDeviceIpv4Addr1 overwrites any union data inside the DeviceIpv4Addr as the provided DeviceIpv4Addr1
func (t *DeviceIpv4Addr) FromDeviceIpv4Addr1(v DeviceIpv4Addr1) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

// MergeDeviceIpv4Addr1 performs a merge with any union data inside the DeviceIpv4Addr, using the provided DeviceIpv4Addr1
func (t *DeviceIpv4Addr) MergeDeviceIpv4Addr1(v DeviceIpv4Addr1) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t DeviceIpv4Addr) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	if err != nil {
		return nil, err
	}
	object := make(map[string]json.RawMessage)
	if t.union != nil {
		err = json.Unmarshal(b, &object)
		if err != nil {
			return nil, err
		}
	}

	if t.PrivateAddress != nil {
		object["privateAddress"], err = json.Marshal(t.PrivateAddress)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'privateAddress': %w", err)
		}
	}

	if t.PublicAddress != nil {
		object["publicAddress"], err = json.Marshal(t.PublicAddress)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'publicAddress': %w", err)
		}
	}

	if t.PublicPort != nil {
		object["publicPort"], err = json.Marshal(t.PublicPort)
		if err != nil {
			return nil, fmt.Errorf("error marshaling 'publicPort': %w", err)
		}
	}
	b, err = json.Marshal(object)
	return b, err
}

func (t *DeviceIpv4Addr) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	object := make(map[string]json.RawMessage)
	err = json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if raw, found := object["privateAddress"]; found {
		err = json.Unmarshal(raw, &t.PrivateAddress)
		if err != nil {
			return fmt.Errorf("error reading 'privateAddress': %w", err)
		}
	}

	if raw, found := object["publicAddress"]; found {
		err = json.Unmarshal(raw, &t.PublicAddress)
		if err != nil {
			return fmt.Errorf("error reading 'publicAddress': %w", err)
		}
	}

	if raw, found := object["publicPort"]; found {
		err = json.Unmarshal(raw, &t.PublicPort)
		if err != nil {
			return fmt.Errorf("error reading 'publicPort': %w", err)
		}
	}

	return err
}
