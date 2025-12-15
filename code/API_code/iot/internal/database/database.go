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
package database

import (
	"context"
	"time"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// Interface defines the database operations for power-saving jobs
type Interface interface {
	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
	GetPendingTransactions(ctx context.Context) ([]*Transaction, error)
	ClaimTransaction(ctx context.Context, transactionID string, action string) (bool, error)
	MarkTransactionFailed(ctx context.Context, transactionID string, errorMsg string) error
	MarkTransactionCompleted(ctx context.Context, transactionID string) error
	CheckDeviceConflicts(ctx context.Context, deviceIDs []string) ([]string, error)
	DeleteOldTransactions(ctx context.Context, olderThan time.Time) (int64, error)

	// Device operations within transaction
	StoreDeviceOriginalState(ctx context.Context, deviceID string, originalState *DeviceOriginalState) error
	GetDeviceOriginalState(ctx context.Context, deviceID string) (*DeviceOriginalState, error)
	CheckDeviceConfigsExist(ctx context.Context, deviceIDs []string) ([]string, error)
	UpdateDeviceActionStatus(ctx context.Context, transactionID string, deviceID string, action string, status string) (allCompleted bool, err error)
	GetTransactionDevices(ctx context.Context, transactionID string, action string) ([]*TransactionDevice, error)
}

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Transaction represents the complete transaction with all devices embedded
type Transaction struct {
	TransactionID       string                     `bson:"_id" json:"transactionId"`
	StartAt             time.Time                  `bson:"startAt" json:"startAt"`
	EndAt               *time.Time                 `bson:"endAt,omitempty" json:"endAt,omitempty"`
	Enabled             bool                       `bson:"enabled" json:"enabled"`
	SubscriptionRequest models.SubscriptionRequest `bson:"subscriptionRequest" json:"subscriptionRequest"`
	Status              Status                     `bson:"status" json:"status"`
	CreatedAt           time.Time                  `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time                  `bson:"updatedAt" json:"updatedAt"`
	ErrorMessage        string                     `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`

	// All devices embedded in the transaction
	Devices []*TransactionDevice `bson:"devices" json:"devices"`

	// Action completion tracking
	StartActionCompleted bool `bson:"startActionCompleted" json:"startActionCompleted"`
	EndActionCompleted   bool `bson:"endActionCompleted" json:"endActionCompleted"`
	StartActionNotified  bool `bson:"startActionNotified" json:"startActionNotified"`
	EndActionNotified    bool `bson:"endActionNotified" json:"endActionNotified"`
}

// TransactionDevice represents a single device within a transaction
type TransactionDevice struct {
	DeviceID    string              `bson:"deviceId" json:"deviceId"`
	Device      models.Device       `bson:"device" json:"device"`
	StartAction *DeviceActionStatus `bson:"startAction,omitempty" json:"startAction,omitempty"`
	EndAction   *DeviceActionStatus `bson:"endAction,omitempty" json:"endAction,omitempty"`
}

// DeviceOriginalState stores the original configuration before actuation in a separate collection
type DeviceOriginalState struct {
	DeviceID              string    `bson:"_id" json:"deviceId"`
	PpMaximumLatency      string    `bson:"ppMaximumLatency" json:"ppMaximumLatency"`
	PpMaximumResponseTime string    `bson:"ppMaximumResponseTime" json:"ppMaximumResponseTime"`
	Timestamp             time.Time `bson:"timestamp" json:"timestamp"`
}

// DeviceActionStatus tracks the status of a device action (start or end)
type DeviceActionStatus struct {
	Status    string    `bson:"status" json:"status"` // "in-progress", "success", "failed"
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}
