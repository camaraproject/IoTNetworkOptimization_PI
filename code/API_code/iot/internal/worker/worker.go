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
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/easyapi"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/event"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

// ActuationWorker handles device actuation requests.
type ActuationWorker struct {
	database     database.Interface
	deviceClient easyapi.Client
	sender       event.Sender
	receiver     event.Receiver
	config       config.PowerSaving
}

// Handler implements receiver.Handler interface for CloudEvents.
type Handler struct {
	worker *ActuationWorker
}

func (h *Handler) Handle(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, error) {
	err := h.worker.handleActuationRequest(ctx, e)
	return nil, err
}

// New creates a new ActuationWorker.
func New(db database.Interface, deviceClient easyapi.Client, sender event.Sender, receiver event.Receiver, powerSavingConfig config.PowerSaving) *ActuationWorker {
	return &ActuationWorker{
		database:     db,
		deviceClient: deviceClient,
		sender:       sender,
		receiver:     receiver,
		config:       powerSavingConfig,
	}
}

// Start begins processing device.actuation.request events.
func (w *ActuationWorker) Start() error {
	log := logger.Get()
	log.Info("Starting actuation worker")

	handler := &Handler{worker: w}
	return w.receiver.Start(handler)
}

// handleActuationRequest processes incoming device.actuation.request events.
func (w *ActuationWorker) handleActuationRequest(ctx context.Context, e cloudevents.Event) error {
	log := logger.Get().With(zap.String("eventId", e.ID()), zap.String("eventType", e.Type()))

	var data event.DeviceActuationRequestData
	if err := json.Unmarshal(e.Data(), &data); err != nil {
		log.Error("Failed to unmarshal actuation data", zap.Error(err))
		return fmt.Errorf("unmarshal actuation data: %w", err)
	}

	deviceID := string(*data.Device.NetworkAccessIdentifier)

	log.Info("Received device actuation request",
		zap.String("deviceId", deviceID),
		zap.String("transactionId", data.TransactionID),
		zap.Bool("enabled", data.Enabled),
		zap.String("action", data.Action))

	if err := w.processDevice(ctx, data.TransactionID, data.Device, deviceID, data.Action, data.Enabled, data.SubscriptionRequest); err != nil {
		log.Error("Failed to process device", zap.Error(err), zap.String("deviceId", deviceID))
		return err
	}

	log.Debug("Device actuation completed successfully", zap.String("deviceId", deviceID))
	return nil
}

// processDevice handles actuation for a single device based on action type.
func (w *ActuationWorker) processDevice(ctx context.Context, transactionID string, device models.Device, deviceID string, action string, enabled bool, subscriptionRequest models.SubscriptionRequest) error {
	log := logger.Get().With(
		zap.String("transactionId", transactionID),
		zap.String("deviceId", deviceID),
		zap.String("action", action),
		zap.Bool("enabled", enabled))

	if _, err := w.database.UpdateDeviceActionStatus(ctx, transactionID, deviceID, action, "in-progress"); err != nil {
		log.Error("Failed to update status to in-progress", zap.Error(err))
		return fmt.Errorf("update status to in-progress: %w", err)
	}

	finalStatus := "success"

	if action == event.ActionStart {
		if enabled {
			log.Debug("Processing start action - applying power-saving", zap.String("deviceId", deviceID))

			currentConfig, err := w.deviceClient.GetDeviceConfig(ctx, device)
			if err != nil {
				log.Error("Failed to get device config", zap.Error(err), zap.String("deviceId", deviceID))
				finalStatus = "failed"
			} else {
				originalState := &database.DeviceOriginalState{
					PpMaximumLatency:      currentConfig.PpMaximumLatency,
					PpMaximumResponseTime: currentConfig.PpMaximumResponseTime,
				}

				if err := w.database.StoreDeviceOriginalState(ctx, deviceID, originalState); err != nil {
					log.Error("Failed to store device original state", zap.Error(err), zap.String("deviceId", deviceID))
					finalStatus = "failed"
				} else {
					log.Debug("Stored original device configuration",
						zap.String("deviceId", deviceID),
						zap.String("ppMaximumLatency", currentConfig.PpMaximumLatency),
						zap.String("ppMaximumResponseTime", currentConfig.PpMaximumResponseTime))

					powerSavingConfig := &easyapi.DeviceConfig{
						PpMaximumLatency:      w.config.MaxLatency,
						PpMaximumResponseTime: w.config.MaxResponseTime,
					}

					if err := w.deviceClient.SetDeviceConfig(ctx, device, powerSavingConfig); err != nil {
						log.Error("Failed to set device config", zap.Error(err), zap.String("deviceId", deviceID))
						finalStatus = "failed"
					} else {
						log.Debug("Device actuation successful - power-saving applied",
							zap.String("deviceId", deviceID))
					}
				}
			}
		} else {
			log.Debug("Processing start action - restoring original config", zap.String("deviceId", deviceID))

			storedState, err := w.database.GetDeviceOriginalState(ctx, deviceID)
			if err != nil {
				log.Error("No original state found for device - cannot restore",
					zap.String("deviceId", deviceID),
					zap.Error(err))
				finalStatus = "failed"
			} else {
				log.Debug("Retrieved original device configuration",
					zap.String("deviceId", deviceID),
					zap.String("ppMaximumLatency", storedState.PpMaximumLatency),
					zap.String("ppMaximumResponseTime", storedState.PpMaximumResponseTime))

				originalConfig := &easyapi.DeviceConfig{
					PpMaximumLatency:      storedState.PpMaximumLatency,
					PpMaximumResponseTime: storedState.PpMaximumResponseTime,
				}

				if err := w.deviceClient.SetDeviceConfig(ctx, device, originalConfig); err != nil {
					log.Error("Failed to restore device config", zap.Error(err), zap.String("deviceId", deviceID))
					finalStatus = "failed"
				} else {
					log.Debug("Device actuation successful - original config restored",
						zap.String("deviceId", deviceID))
				}
			}
		}

	} else if action == event.ActionEnd {
		if enabled {
			log.Debug("Processing end action - restoring original config", zap.String("deviceId", deviceID))

			storedState, err := w.database.GetDeviceOriginalState(ctx, deviceID)
			if err != nil {
				log.Error("No original state found for device - START action likely failed",
					zap.String("deviceId", deviceID),
					zap.Error(err))
				finalStatus = "failed"
			} else {
				log.Debug("Retrieved original device configuration",
					zap.String("deviceId", deviceID),
					zap.String("ppMaximumLatency", storedState.PpMaximumLatency),
					zap.String("ppMaximumResponseTime", storedState.PpMaximumResponseTime))

				originalConfig := &easyapi.DeviceConfig{
					PpMaximumLatency:      storedState.PpMaximumLatency,
					PpMaximumResponseTime: storedState.PpMaximumResponseTime,
				}

				if err := w.deviceClient.SetDeviceConfig(ctx, device, originalConfig); err != nil {
					log.Error("Failed to restore device config", zap.Error(err), zap.String("deviceId", deviceID))
					finalStatus = "failed"
				} else {
					log.Debug("Device actuation successful - original config restored",
						zap.String("deviceId", deviceID))
				}
			}
		} else {
			log.Debug("Processing end action - applying power-saving", zap.String("deviceId", deviceID))

			powerSavingConfig := &easyapi.DeviceConfig{
				PpMaximumLatency:      w.config.MaxLatency,
				PpMaximumResponseTime: w.config.MaxResponseTime,
			}

			if err := w.deviceClient.SetDeviceConfig(ctx, device, powerSavingConfig); err != nil {
				log.Error("Failed to set device config", zap.Error(err), zap.String("deviceId", deviceID))
				finalStatus = "failed"
			} else {
				log.Debug("Device actuation successful - power-saving applied",
					zap.String("deviceId", deviceID))
			}
		}
	}

	allComplete, err := w.database.UpdateDeviceActionStatus(ctx, transactionID, deviceID, action, finalStatus)
	if err != nil {
		log.Error("Failed to update device status", zap.Error(err))
		return fmt.Errorf("update device status: %w", err)
	}

	log.Debug("Device status updated",
		zap.String("deviceId", deviceID),
		zap.String("status", finalStatus),
		zap.Bool("allComplete", allComplete))

	if allComplete {
		log.Info("All devices completed, sending notification event",
			zap.String("transactionId", transactionID),
			zap.String("action", action))

		allCompletedData := event.AllDevicesCompletedData{
			TransactionID:       transactionID,
			Action:              action,
			CompletedAt:         time.Now(),
			SubscriptionRequest: subscriptionRequest,
		}

		eventID := fmt.Sprintf("%s-%s-all-completed", transactionID, action)
		if err := w.sender.Send(ctx, eventID, event.EventTypeAllDevicesCompleted, event.SourceiotWorker, allCompletedData); err != nil {
			log.Error("Failed to send all-devices.completed event", zap.Error(err))
			return fmt.Errorf("send all-devices.completed event: %w", err)
		}

		log.Debug("All-devices.completed event sent successfully")

		if err := w.markTransactionCompleteIfDone(ctx, transactionID, action); err != nil {
			log.Error("Failed to mark transaction as completed", zap.Error(err))
		}
	}

	return nil
}

// markTransactionCompleteIfDone marks the transaction as completed when all actions finish.
func (w *ActuationWorker) markTransactionCompleteIfDone(ctx context.Context, transactionID string, action string) error {
	log := logger.Get().With(zap.String("transactionId", transactionID), zap.String("action", action))

	transaction, err := w.database.GetTransaction(ctx, transactionID)
	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err))
		return err
	}

	shouldComplete := false
	if action == event.ActionStart && transaction.EndAt == nil {
		shouldComplete = true
		log.Debug("START action completed and no END scheduled - marking transaction as completed")
	} else if action == event.ActionEnd {
		shouldComplete = true
		log.Debug("END action completed - marking transaction as completed")
	}

	if shouldComplete {
		if err := w.database.MarkTransactionCompleted(ctx, transactionID); err != nil {
			log.Error("Failed to mark transaction as completed", zap.Error(err))
			return err
		}
		log.Info("Transaction marked as completed")
	}

	return nil
}
