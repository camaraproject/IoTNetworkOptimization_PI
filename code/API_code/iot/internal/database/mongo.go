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
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

var _ Interface = &mongoDB{}

type mongoDB struct {
	transactions  *mongo.Collection
	deviceConfigs *mongo.Collection
}

// NewMongoDB creates a new MongoDB connection using the provided URI and database name.
func NewMongoDB(conf config.Database) (Interface, error) {
	clientOpts := options.Client().ApplyURI(conf.Uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}
	db := client.Database(conf.Name)
	transactionsColl := db.Collection("transactions")
	deviceConfigsColl := db.Collection("device_configs")

	return &mongoDB{
		transactions:  transactionsColl,
		deviceConfigs: deviceConfigsColl,
	}, nil
}

// CreateTransaction creates a new transaction document with embedded devices.
func (m *mongoDB) CreateTransaction(ctx context.Context, transaction *Transaction) error {
	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	if transaction.Devices == nil {
		transaction.Devices = []*TransactionDevice{}
	}

	_, err := m.transactions.InsertOne(ctx, transaction)
	return err
}

// GetTransaction retrieves a single transaction by ID.
func (m *mongoDB) GetTransaction(ctx context.Context, transactionID string) (*Transaction, error) {
	var transaction Transaction
	err := m.transactions.FindOne(ctx, bson.M{"_id": transactionID}).Decode(&transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetPendingTransactions retrieves all transactions that are pending or processing.
func (m *mongoDB) GetPendingTransactions(ctx context.Context) ([]*Transaction, error) {
	log := logger.Get()

	filter := bson.M{
		"status": bson.M{"$in": []Status{StatusPending, StatusProcessing}},
	}

	cursor, err := m.transactions.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query pending transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []*Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, fmt.Errorf("decode pending transactions: %w", err)
	}

	log.Info("Retrieved pending transactions", zap.Int("count", len(transactions)))
	return transactions, nil
}

// ClaimTransaction atomically claims a transaction for a specific action.
func (m *mongoDB) ClaimTransaction(ctx context.Context, transactionID string, action string) (bool, error) {
	filter := bson.M{"_id": transactionID}

	if action == "start" {
		filter["startActionCompleted"] = bson.M{"$ne": true}
	} else if action == "end" {
		filter["endActionCompleted"] = bson.M{"$ne": true}
	}

	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	res, err := m.transactions.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}

	// If no documents matched, the transaction doesn't exist or action is already completed
	if res.MatchedCount == 0 {
		// Try to fetch the transaction to distinguish between "not found" and "already completed"
		var transaction Transaction
		err := m.transactions.FindOne(ctx, bson.M{"_id": transactionID}).Decode(&transaction)
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("transaction not found: %s", transactionID)
		}
		if err != nil {
			return false, err
		}
		// Transaction exists but action already completed
		return false, nil
	}

	return true, nil
}

// MarkTransactionFailed marks a transaction as failed.
func (m *mongoDB) MarkTransactionFailed(ctx context.Context, transactionID string, errorMsg string) error {
	filter := bson.M{"_id": transactionID}
	update := bson.M{
		"$set": bson.M{
			"status":       StatusFailed,
			"errorMessage": errorMsg,
			"updatedAt":    time.Now(),
		},
	}
	_, err := m.transactions.UpdateOne(ctx, filter, update)
	return err
}

// MarkTransactionCompleted marks a transaction as completed.
func (m *mongoDB) MarkTransactionCompleted(ctx context.Context, transactionID string) error {
	filter := bson.M{"_id": transactionID}
	update := bson.M{
		"$set": bson.M{
			"status":    StatusCompleted,
			"updatedAt": time.Now(),
		},
	}
	_, err := m.transactions.UpdateOne(ctx, filter, update)
	return err
}

// StoreDeviceOriginalState stores the original device configuration.
func (m *mongoDB) StoreDeviceOriginalState(ctx context.Context, deviceID string, originalState *DeviceOriginalState) error {
	originalState.DeviceID = deviceID
	originalState.Timestamp = time.Now()

	opts := options.Replace().SetUpsert(true)
	_, err := m.deviceConfigs.ReplaceOne(ctx, bson.M{"_id": deviceID}, originalState, opts)
	return err
}

// GetDeviceOriginalState retrieves the original device configuration.
func (m *mongoDB) GetDeviceOriginalState(ctx context.Context, deviceID string) (*DeviceOriginalState, error) {
	var state DeviceOriginalState
	err := m.deviceConfigs.FindOne(ctx, bson.M{"_id": deviceID}).Decode(&state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// CheckDeviceConfigsExist returns deviceIDs that don't have stored configurations.
func (m *mongoDB) CheckDeviceConfigsExist(ctx context.Context, deviceIDs []string) ([]string, error) {
	log := logger.Get()
	missing := make([]string, 0)

	for _, deviceID := range deviceIDs {
		count, err := m.deviceConfigs.CountDocuments(ctx, bson.M{"_id": deviceID})
		if err != nil {
			log.Error("Failed to check device config existence",
				zap.String("deviceId", deviceID),
				zap.Error(err))
			return nil, fmt.Errorf("check device config existence: %w", err)
		}
		if count == 0 {
			missing = append(missing, deviceID)
		}
	}

	return missing, nil
}

// UpdateDeviceActionStatus updates the status of a device action.
// Returns true if all devices are complete and this caller won the notification race.
func (m *mongoDB) UpdateDeviceActionStatus(ctx context.Context, transactionID string, deviceID string, action string, status string) (allCompleted bool, err error) {
	actionField := "devices.$.startAction"
	notifiedField := "startActionNotified"
	if action == "end" {
		actionField = "devices.$.endAction"
		notifiedField = "endActionNotified"
	}

	filter := bson.M{
		"_id":              transactionID,
		"devices.deviceId": deviceID,
	}
	update := bson.M{
		"$set": bson.M{
			actionField: &DeviceActionStatus{
				Status:    status,
				Timestamp: time.Now(),
			},
			"updatedAt": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var transaction Transaction
	err = m.transactions.FindOneAndUpdate(ctx, filter, update, opts).Decode(&transaction)
	if err != nil {
		return false, err
	}

	completedCount := 0
	totalDevices := len(transaction.Devices)

	for _, device := range transaction.Devices {
		var deviceAction *DeviceActionStatus
		if action == "start" {
			deviceAction = device.StartAction
		} else {
			deviceAction = device.EndAction
		}

		if deviceAction != nil && (deviceAction.Status == "success" || deviceAction.Status == "failed") {
			completedCount++
		}
	}

	// Mark as notified atomically to prevent duplicate notifications.
	if completedCount == totalDevices && totalDevices > 0 {
		filter := bson.M{
			"_id":         transactionID,
			notifiedField: false,
		}
		update := bson.M{
			"$set": bson.M{
				notifiedField: true,
				"updatedAt":   time.Now(),
			},
		}
		result, err := m.transactions.UpdateOne(ctx, filter, update)
		if err != nil {
			return false, err
		}

		return result.MatchedCount == 1, nil
	}

	return false, nil
}

// GetTransactionDevices retrieves all devices for a transaction with their action status.
func (m *mongoDB) GetTransactionDevices(ctx context.Context, transactionID string, action string) ([]*TransactionDevice, error) {
	transaction, err := m.GetTransaction(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	return transaction.Devices, nil
}

// CheckDeviceConflicts returns transactionIDs of active transactions containing the specified devices.
func (m *mongoDB) CheckDeviceConflicts(ctx context.Context, deviceIDs []string) ([]string, error) {
	filter := bson.M{
		"status": bson.M{
			"$in": []Status{StatusPending, StatusProcessing},
		},
		"devices.deviceId": bson.M{
			"$in": deviceIDs,
		},
	}

	cursor, err := m.transactions.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conflictingTxs []string
	for cursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		conflictingTxs = append(conflictingTxs, result.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return conflictingTxs, nil
}

// DeleteOldTransactions removes completed or failed transactions older than the specified time.
func (m *mongoDB) DeleteOldTransactions(ctx context.Context, olderThan time.Time) (int64, error) {
	log := logger.Get()

	filter := bson.M{
		"status": bson.M{
			"$in": []Status{StatusCompleted, StatusFailed},
		},
		"updatedAt": bson.M{
			"$lt": olderThan,
		},
	}

	result, err := m.transactions.DeleteMany(ctx, filter)
	if err != nil {
		log.Error("Failed to delete old transactions", zap.Error(err))
		return 0, fmt.Errorf("delete old transactions: %w", err)
	}

	if result.DeletedCount > 0 {
		log.Info("Deleted old transactions",
			zap.Int64("count", result.DeletedCount),
			zap.Time("olderThan", olderThan))
	}

	return result.DeletedCount, nil
}
