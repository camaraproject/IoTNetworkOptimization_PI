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
package event

import (
	"time"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// Action constants for device actuation.
const (
	ActionStart = "start" // Trigger at schedule start
	ActionEnd   = "end"   // Trigger at schedule end
)

// ScheduleRequestedData is the payload for schedule.requested events.
type ScheduleRequestedData struct {
	StartAt time.Time          `json:"startAt"`
	EndAt   *time.Time         `json:"endAt,omitempty"`
	Payload PowerSavingPayload `json:"payload"`
}

// PowerSavingPayload contains the device actuation data.
type PowerSavingPayload struct {
	Devices             []models.Device            `json:"devices"`
	Enabled             bool                       `json:"enabled"`
	SubscriptionRequest models.SubscriptionRequest `json:"subscriptionRequest"`
	TransactionID       string                     `json:"transactionId"`
}

// DeviceActuationRequestData is the payload for device.actuation.request events.
type DeviceActuationRequestData struct {
	Device              models.Device              `json:"device"`
	Enabled             bool                       `json:"enabled"`
	TransactionID       string                     `json:"transactionId"`
	Action              string                     `json:"action"` // "start" or "end" (use ActionStart/ActionEnd constants)
	SubscriptionRequest models.SubscriptionRequest `json:"subscriptionRequest"`
}

// AllDevicesCompletedData is the payload for all-devices.completed events.
type AllDevicesCompletedData struct {
	TransactionID       string                     `json:"transactionId"`
	Action              string                     `json:"action"` // "start" or "end"
	CompletedAt         time.Time                  `json:"completedAt"`
	SubscriptionRequest models.SubscriptionRequest `json:"subscriptionRequest"`
}

// ErrorNotificationData is the payload for power-saving.error events.
type ErrorNotificationData struct {
	TransactionID       string                     `json:"transactionId"`
	Status              int                        `json:"status"`
	Code                string                     `json:"code"`
	Message             string                     `json:"message"`
	Action              string                     `json:"action,omitempty"` // "start" or "end" if applicable
	SubscriptionRequest models.SubscriptionRequest `json:"subscriptionRequest"`
}
