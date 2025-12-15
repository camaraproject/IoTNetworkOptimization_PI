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
package notifier

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/event"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

// NotificationWorker handles notification callbacks.
type NotificationWorker struct {
	database database.Interface
	receiver event.Receiver
	config   config.HTTP
}

// Handler implements receiver.Handler interface for CloudEvents.
type Handler struct {
	worker *NotificationWorker
}

func (h *Handler) Handle(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, error) {
	// Route to appropriate handler based on event type
	switch e.Type() {
	case string(event.EventTypeAllDevicesCompleted):
		return nil, h.worker.handleAllDevicesCompleted(ctx, e)
	case string(event.EventTypePowerSavingError):
		return nil, h.worker.handleErrorNotification(ctx, e)
	default:
		logger.Get().Warn("Unknown event type received", zap.String("eventType", e.Type()))
		return nil, nil
	}
}

// New creates a new NotificationWorker.
func New(db database.Interface, receiver event.Receiver) *NotificationWorker {
	cfg := config.GetConf()
	log := logger.Get()

	if cfg.HTTP.InsecureSkipVerify {
		log.Warn("HTTP_INSECURE_SKIP_VERIFY enabled - TLS verification disabled for internal cluster services")
	}

	return &NotificationWorker{
		database: db,
		receiver: receiver,
		config:   cfg.HTTP,
	}
}

// getHTTPClient returns an HTTP client configured based on the sink URL.
// For internal cluster services (*.svc.cluster.local), TLS verification
// can be skipped if configured via HTTP_INSECURE_SKIP_VERIFY.
func (w *NotificationWorker) getHTTPClient(sinkURL string) *http.Client {
	// Check if the sink is an internal cluster service
	if isInternalClusterService(sinkURL) {
		logger.Get().Debug("Detected internal cluster service", zap.String("sink", sinkURL), zap.Bool("insecureSkipVerify", w.config.InsecureSkipVerify))
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: w.config.InsecureSkipVerify,
				},
			},
		}
	}

	// For external services, use standard client with system CA pool
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// isInternalClusterService checks if the URL points to an internal Kubernetes service.
// Returns true for URLs with hostnames ending in .svc.cluster.local or just .svc
func isInternalClusterService(sinkURL string) bool {
	u, err := url.Parse(sinkURL)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(u.Hostname())
	return strings.HasSuffix(hostname, ".svc.cluster.local") ||
		strings.HasSuffix(hostname, ".svc") ||
		strings.Contains(hostname, ".svc.")
}

// Start begins processing all-devices.completed events.
func (w *NotificationWorker) Start() error {
	log := logger.Get()
	log.Info("Starting notification worker")

	handler := &Handler{worker: w}
	return w.receiver.Start(handler)
}

// handleAllDevicesCompleted processes incoming all-devices.completed events.
func (w *NotificationWorker) handleAllDevicesCompleted(ctx context.Context, ce cloudevents.Event) error {
	log := logger.Get().With(zap.String("eventId", ce.ID()), zap.String("eventType", ce.Type()))
	log.Debug("Received all-devices.completed event")

	// Parse event data
	var data event.AllDevicesCompletedData
	if err := json.Unmarshal(ce.Data(), &data); err != nil {
		log.Error("Failed to parse event data", zap.Error(err))
		return fmt.Errorf("failed to parse event data: %w", err)
	}

	log = log.With(zap.String("transactionID", data.TransactionID), zap.String("action", data.Action))
	log.Debug("Processing notification callback")

	// Check if subscription request has a sink
	if data.SubscriptionRequest.Sink == "" {
		log.Warn("No notification sink configured, skipping notification")
		return nil
	}

	// Send callback for both START and END actions
	log.Info("Sending callback notification for action completion",
		zap.String("action", data.Action))

	// Get all device results for this action
	devices, err := w.database.GetTransactionDevices(ctx, data.TransactionID, data.Action)
	if err != nil {
		log.Error("Failed to get transaction devices", zap.Error(err))
		return fmt.Errorf("failed to get transaction devices: %w", err)
	}

	log.Debug("Retrieved transaction devices", zap.Int("deviceCount", len(devices)))

	// Build activation status array
	activationStatus := make([]models.DeviceStatus, 0, len(devices))
	for _, txDevice := range devices {
		// Get the appropriate action status
		var actionStatus *database.DeviceActionStatus
		if data.Action == event.ActionStart {
			actionStatus = txDevice.StartAction
		} else {
			actionStatus = txDevice.EndAction
		}

		// Convert status string to DeviceStatusStatus
		var status models.DeviceStatusStatus
		if actionStatus != nil {
			switch actionStatus.Status {
			case "success":
				status = models.Success
			case "failed":
				status = models.Failed
			case "in-progress":
				status = models.InProgress
			case "pending":
				status = models.Pending
			default:
				log.Warn("Unknown status, defaulting to failed", zap.String("status", actionStatus.Status))
				status = models.Failed
			}
		} else {
			// No action status found, default to pending
			status = models.Pending
		}

		activationStatus = append(activationStatus, models.DeviceStatus{
			Device: &txDevice.Device,
			Status: &status,
		})
	}

	// Build PowerSavingResponse
	transactionID := data.TransactionID
	response := models.PowerSavingResponse{
		ActivationStatus: &activationStatus,
		TransactionId:    &transactionID,
	}

	// Create CloudEvent
	notifEvent := cloudevents.NewEvent()
	notifEvent.SetID(data.TransactionID + "-" + data.Action)
	notifEvent.SetSource(string(event.SourceiotAPI))
	notifEvent.SetType(string(models.EventTypeNotificationOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSaving))
	notifEvent.SetTime(time.Now())
	if err := notifEvent.SetData(cloudevents.ApplicationJSON, response); err != nil {
		log.Error("Failed to set CloudEvent data", zap.Error(err))
		return fmt.Errorf("failed to set CloudEvent data: %w", err)
	}

	// Marshal CloudEvent to JSON
	responseBytes, err := json.Marshal(notifEvent)
	if err != nil {
		log.Error("Failed to marshal CloudEvent", zap.Error(err))
		return fmt.Errorf("failed to marshal CloudEvent: %w", err)
	}

	// Build HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", data.SubscriptionRequest.Sink, bytes.NewBuffer(responseBytes))
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/cloudevents+json")

	// Add authorization if credential provided
	if authHeader, ok := data.SubscriptionRequest.SinkCredential.AuthorizationHeader(); ok {
		req.Header.Set("Authorization", authHeader)
		log.Debug("Added authorization header")
	}

	log.Info("Sending callback notification", zap.String("url", data.SubscriptionRequest.Sink))
	client := w.getHTTPClient(data.SubscriptionRequest.Sink)
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send callback notification", zap.Error(err))
		// Don't return error - accept failure and let Knative move on
		return nil
	}
	defer resp.Body.Close()

	log.Debug("Callback notification sent", zap.Int("statusCode", resp.StatusCode))

	// Handle response status codes per CAMARA spec
	switch resp.StatusCode {
	case http.StatusAccepted, http.StatusNoContent, http.StatusOK:
		// 202: Data received successfully
		// 204: No longer interested in updates
		// 200: Success
		log.Info("Callback notification completed successfully", zap.Int("statusCode", resp.StatusCode))
		return nil
	case http.StatusGone:
		// 410: Callback endpoint is gone - stop retrying
		log.Warn("Callback endpoint is gone (410), notification discarded")
		return nil
	case http.StatusTooManyRequests:
		// 429: Too many requests - could implement backoff
		log.Warn("Too many requests (429), notification may need retry")
		return nil
	default:
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Client errors - don't retry
			log.Warn("Callback notification received client error", zap.Int("statusCode", resp.StatusCode))
			return nil
		}
		if resp.StatusCode >= 500 {
			// Server errors - could retry but we accept failure
			log.Warn("Callback notification received server error", zap.Int("statusCode", resp.StatusCode))
			return nil
		}
		log.Warn("Callback notification received unexpected status", zap.Int("statusCode", resp.StatusCode))
		return nil
	}
}

// handleErrorNotification processes error notification events and sends them to the consumer.
func (w *NotificationWorker) handleErrorNotification(ctx context.Context, ce cloudevents.Event) error {
	log := logger.Get().With(zap.String("eventId", ce.ID()), zap.String("eventType", ce.Type()))
	log.Debug("Received error notification event")

	// Parse event data
	var errorData event.ErrorNotificationData
	if err := json.Unmarshal(ce.Data(), &errorData); err != nil {
		log.Error("Failed to parse error event data", zap.Error(err))
		return fmt.Errorf("failed to parse error event data: %w", err)
	}

	log = log.With(
		zap.String("transactionID", errorData.TransactionID),
		zap.String("errorCode", errorData.Code),
		zap.String("action", errorData.Action))
	log.Info("Processing error notification callback")

	// Check if subscription request has a sink
	if errorData.SubscriptionRequest.Sink == "" {
		log.Warn("No notification sink configured, skipping error notification")
		return nil
	}

	// Create error CloudEvent
	notifEvent := cloudevents.NewEvent()
	notifEvent.SetID(errorData.TransactionID + "-error")
	notifEvent.SetSource(string(event.SourceiotNotify))
	notifEvent.SetType(string(models.EventTypeNotificationOrgCamaraprojectIotNetworkOptimizationNotificationV1PowerSavingError))
	notifEvent.SetTime(time.Now())

	// Build error data payload matching CloudEventError schema
	errorPayload := map[string]interface{}{
		"transactionId": errorData.TransactionID,
		"status":        errorData.Status,
		"code":          errorData.Code,
		"message":       errorData.Message,
	}

	if err := notifEvent.SetData(cloudevents.ApplicationJSON, errorPayload); err != nil {
		log.Error("Failed to set CloudEvent data", zap.Error(err))
		return fmt.Errorf("failed to set CloudEvent data: %w", err)
	}

	// Marshal CloudEvent to JSON
	responseBytes, err := json.Marshal(notifEvent)
	if err != nil {
		log.Error("Failed to marshal CloudEvent", zap.Error(err))
		return fmt.Errorf("failed to marshal CloudEvent: %w", err)
	}

	log.Debug("CloudEvent error notification", zap.String("payload", string(responseBytes)))

	// Build HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", errorData.SubscriptionRequest.Sink, bytes.NewBuffer(responseBytes))
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/cloudevents+json")

	// Add authorization if credential provided
	if authHeader, ok := errorData.SubscriptionRequest.SinkCredential.AuthorizationHeader(); ok {
		req.Header.Set("Authorization", authHeader)
		log.Debug("Added authorization header")
	}

	log.Info("Sending error notification callback", zap.String("url", errorData.SubscriptionRequest.Sink))
	client := w.getHTTPClient(errorData.SubscriptionRequest.Sink)
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send error notification callback", zap.Error(err))
		// Don't return error - accept failure and let Knative move on
		return nil
	}
	defer resp.Body.Close()

	log.Debug("Error notification callback sent", zap.Int("statusCode", resp.StatusCode))

	// Handle response status codes per CAMARA spec
	switch resp.StatusCode {
	case http.StatusAccepted, http.StatusNoContent, http.StatusOK:
		log.Info("Error notification callback completed successfully", zap.Int("statusCode", resp.StatusCode))
		return nil
	case http.StatusGone:
		log.Warn("Callback endpoint is gone (410), error notification discarded")
		return nil
	default:
		log.Warn("Error notification callback received unexpected status", zap.Int("statusCode", resp.StatusCode))
		return nil
	}
}
