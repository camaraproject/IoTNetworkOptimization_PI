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
	"fmt"
)

// Manual polymorphic types implementation.
//
// This file consolidates manual implementations for discriminator-based schemas
// excluded from OpenAPI code generation:
//   - SinkCredential (discriminator: credentialType)
//   - SubscriptionRequest (discriminator: protocol)
// The upstream spec uses OpenAPI discriminators with mappings, but the current
// oapi-codegen version does not emit ergonomic polymorphic Go types that allow
// dynamic decoding into a single wrapper. We therefore exclude these schemas
// (see api/config/models.yaml) and implement custom wrappers.
//
// Currently, only the following variants are supported:
//   - SinkCredential with credentialType = "ACCESSTOKEN" (provides bearer token elsewhere in spec)
//   - SubscriptionRequest with protocol = "HTTP"
//
// If future generator releases support these discriminators natively, this file
// can be removed and exclusion entries deleted.

// -----------------------------------------------------------------------------
// SinkCredential (credentialType discriminator)
// -----------------------------------------------------------------------------

// SubscriptionRequest The request for creating a event-type event subscription
type SubscriptionRequest struct {
	// Config Implementation-specific configuration parameters needed by the subscription manager for acquiring events.
	// In CAMARA we have predefined attributes like `subscriptionExpireTime`, `subscriptionMaxEvents`, `initialEvent`
	// Specific event type attributes must be defined in `subscriptionDetail`
	// Note: if a request is performed for several event type, all subscribed event will use same `config` parameters.
	Config Config `json:"config" bson:"config"`

	// Protocol Identifier of a delivery protocol. Only HTTP is allowed for now
	Protocol         Protocol      `json:"protocol" bson:"protocol"`
	ProtocolSettings *HTTPSettings `json:"protocolSettings,omitempty" bson:"protocolSettings,omitempty"`

	// Sink The address to which events shall be delivered using the selected protocol.
	Sink string `json:"sink" bson:"sink"`

	// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
	SinkCredential *SinkCredential `json:"sinkCredential,omitempty" bson:"sinkCredential,omitempty"`

	// Types Camara Event types eligible to be delivered by this subscription.
	// Note: the maximum number of event types per subscription will be decided at API project level
	Types []SubscriptionEventType `json:"types" bson:"types"`
}

// SinkCredential A sink credential provides authentication or authorization information necessary to enable delivery of events to a target.
type SinkCredential struct {
	// CredentialType The type of the credential.
	// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
	CredentialType SinkCredentialCredentialType `json:"credentialType"`
	AccessTokenCredential
}

// SinkCredentialCredentialType The type of the credential.
// Note: Type of the credential - MUST be set to ACCESSTOKEN for now
type SinkCredentialCredentialType string

const (
	SinkCredentialCredentialTypeACCESSTOKEN  SinkCredentialCredentialType = "ACCESSTOKEN"
	SinkCredentialCredentialTypePLAIN        SinkCredentialCredentialType = "PLAIN"
	SinkCredentialCredentialTypeREFRESHTOKEN SinkCredentialCredentialType = "REFRESHTOKEN"
)

// Validate enforces only ACCESSTOKEN is currently supported and required fields present.
func (sc *SinkCredential) Validate() error {
	if sc == nil {
		return nil
	}
	if sc.CredentialType != SinkCredentialCredentialTypeACCESSTOKEN {
		return fmt.Errorf("sink credential type '%s' not implemented (only ACCESSTOKEN supported)", sc.CredentialType)
	}
	return nil
}

// AuthorizationHeader builds Authorization header value if valid.
func (sc *SinkCredential) AuthorizationHeader() (string, bool) {
	if sc == nil {
		return "", false
	}
	if sc.CredentialType != SinkCredentialCredentialTypeACCESSTOKEN {
		return "", false
	}
	if sc.AccessToken == "" || sc.AccessTokenType != "bearer" {
		return "", false
	}
	return "Bearer " + sc.AccessToken, true
}

// ValidateProtocol enforces only HTTP protocol supported in subscription.
func (sr *SubscriptionRequest) ValidateProtocol() error {
	if sr.Protocol != HTTP {
		return fmt.Errorf("subscription protocol '%s' not implemented; only HTTP supported", sr.Protocol)
	}
	return nil
}

// ValidateTypes ensures the subscription types array contains both required event types.
func (sr *SubscriptionRequest) ValidateTypes() error {
	if len(sr.Types) != 2 {
		return fmt.Errorf("subscription types must contain exactly 2 event types ('%s' and '%s'), got %d",
			SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving,
			SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError,
			len(sr.Types))
	}

	hasPowerSaving := false
	hasPowerSavingError := false

	for _, eventType := range sr.Types {
		switch eventType {
		case SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving:
			hasPowerSaving = true
		case SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError:
			hasPowerSavingError = true
		}
	}

	if !hasPowerSaving || !hasPowerSavingError {
		return fmt.Errorf("subscription types must contain both '%s' and '%s'",
			SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving,
			SubscriptionEventTypeOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError,
		)
	}

	return nil
}
