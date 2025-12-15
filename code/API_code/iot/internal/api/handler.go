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
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/server"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/deviceidentifier"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/event"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

var _ server.ServerInterface = &handler{}

func New(db database.Interface) (*handler, error) {
	sender, err := event.NewSender()
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud event sender: %w", err)
	}
	return &handler{
		events:     sender,
		database:   db,
		translator: deviceidentifier.NewMockTranslator(),
	}, nil
}

type handler struct {
	database   database.Interface
	events     event.Sender
	translator deviceidentifier.Translator
}

// ActivatePowerSaving implements server.ServerInterface.
// It validates the request, creates a stable scheduleId, publishes a schedule.requested event, and returns 202 Accepted immediately.
func (h *handler) ActivatePowerSaving(ctx echo.Context, params models.ActivatePowerSavingParams) error {
	log := logger.Get()

	var req models.PowerSavingRequest
	if err := ctx.Bind(&req); err != nil {
		log.Error("Failed to bind request", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: "invalid request body",
		})
	}

	// Validate format fields (IPv4, IPv6) or types with AllOf/AnyOf
	if err := validatePowerSavingRequest(&req); err != nil {
		log.Error("Format validation failed", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: fmt.Sprintf("format validation error: %v", err),
		})
	}

	// Validate protocol
	if err := req.SubscriptionRequest.ValidateProtocol(); err != nil {
		log.Error("Invalid subscription protocol", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: fmt.Sprintf("invalid subscription protocol: %v", err),
		})
	}

	// Validate subscription types
	if err := req.SubscriptionRequest.ValidateTypes(); err != nil {
		log.Error("Invalid subscription types", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: fmt.Sprintf("invalid subscription types: %v", err),
		})
	}

	// Validate sink credential
	if err := req.SubscriptionRequest.SinkCredential.Validate(); err != nil {
		log.Error("Invalid sink credential", zap.Error(err))
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: fmt.Sprintf("invalid sink credential: %v", err),
		})
	}

	// Validate request
	if len(req.Devices) == 0 {
		return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ARGUMENT",
			Message: "devices list cannot be empty",
		})
	}

	// Resolve all device identifiers to networkAccessIdentifier and check for duplicates
	deviceIDs := make([]string, 0, len(req.Devices))
	seen := make(map[string]bool)
	for i := range req.Devices {
		// Resolve device identifiers to networkAccessIdentifier
		nai, err := h.translator.ResolveNetworkAccessIdentifier(ctx.Request().Context(), req.Devices[i])
		if err != nil {
			log.Error("Failed to resolve device identifier",
				zap.Int("deviceIndex", i),
				zap.Error(err))
			return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_ARGUMENT",
				Message: fmt.Sprintf("failed to resolve device identifier at index %d: %v", i, err),
			})
		}

		// Set the resolved networkAccessIdentifier on the device
		req.Devices[i].NetworkAccessIdentifier = &nai

		deviceID := string(nai)

		// Check for duplicates
		if seen[deviceID] {
			return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_ARGUMENT",
				Message: fmt.Sprintf("duplicate device in request: %s", deviceID),
			})
		}
		seen[deviceID] = true
		deviceIDs = append(deviceIDs, deviceID)
	}

	// If enabled is false, verify all devices have stored configurations
	if !req.Enabled {
		missingConfigs, err := h.database.CheckDeviceConfigsExist(ctx.Request().Context(), deviceIDs)
		if err != nil {
			log.Error("Failed to check device configs existence", zap.Error(err))
			return ctx.JSON(http.StatusInternalServerError, models.ErrorInfo{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL",
				Message: "failed to verify device configurations",
			})
		}

		if len(missingConfigs) > 0 {
			log.Warn("Cannot disable power-saving: missing stored configurations",
				zap.Strings("missingDevices", missingConfigs))
			return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_ARGUMENT",
				Message: fmt.Sprintf("cannot disable power-saving: no stored configuration for devices: %v", missingConfigs),
			})
		}
	}
	// Check for conflicting transactions
	conflicts, err := h.database.CheckDeviceConflicts(ctx.Request().Context(), deviceIDs)
	if err != nil {
		log.Error("Failed to check device conflicts", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, models.ErrorInfo{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL",
			Message: "failed to check device availability",
		})
	}

	if len(conflicts) > 0 {
		log.Warn("Request contains devices already in active transactions",
			zap.Strings("conflictingTransactions", conflicts),
			zap.Int("conflictCount", len(conflicts)))
		return ctx.JSON(http.StatusConflict, models.ErrorInfo{
			Status:  http.StatusConflict,
			Code:    "CONFLICT",
			Message: fmt.Sprintf("one or more devices are already in use by active transactions: %v", conflicts),
		})
	}

	// Generate stable scheduleId from request content
	transactionID := uuid.New().String()

	var startAt time.Time
	var endAt *time.Time

	if req.TimePeriod != nil {
		startAt = req.TimePeriod.StartDate
		if req.TimePeriod.EndDate != nil {
			endAt = req.TimePeriod.EndDate
			// Validate that endDate is after startDate
			if !endAt.After(startAt) {
				return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
					Status:  http.StatusBadRequest,
					Code:    "INVALID_ARGUMENT",
					Message: "endDate must be after startDate",
				})
			}
		}

		// Validate dates in the past
		now := time.Now()
		startInPast := startAt.Before(now)
		endInPast := endAt != nil && endAt.Before(now)

		if startInPast && endInPast {
			// Both dates in the past - reject transaction
			log.Warn("Transaction with both dates in the past rejected",
				zap.Time("startAt", startAt),
				zap.Time("endAt", *endAt))
			return ctx.JSON(http.StatusBadRequest, models.ErrorInfo{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_ARGUMENT",
				Message: "both startDate and endDate are in the past",
			})
		}
		// If startDate is in the past but endDate is not, we proceed with the transaction
	} else {
		// If no time period, execute immediately
		startAt = time.Now()
	}

	// Prepare schedule.requested event payload
	scheduleData := event.ScheduleRequestedData{
		StartAt: startAt,
		EndAt:   endAt,
		Payload: event.PowerSavingPayload{
			Devices:             req.Devices,
			Enabled:             req.Enabled,
			TransactionID:       transactionID,
			SubscriptionRequest: req.SubscriptionRequest,
		},
	}

	err = h.events.Send(
		ctx.Request().Context(),
		transactionID,
		event.EventTypeScheduleRequested,
		event.SourceiotAPI,
		scheduleData,
	)
	if err != nil {
		log.Error("Failed to send schedule.requested event", zap.Error(err), zap.String("transactionId", transactionID))
		return ctx.JSON(http.StatusInternalServerError, models.ErrorInfo{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL",
			Message: "failed to schedule request",
		})
	}

	log.Info("Schedule requested",
		zap.String("transactionId", transactionID),
		zap.Time("startAt", startAt))

	// Return 202 Accepted with transaction ID
	response := models.PowerSavingResponse{
		TransactionId: &transactionID,
	}

	return ctx.JSON(http.StatusAccepted, response)
}

// GetPowerSaving implements server.ServerInterface.
func (h *handler) GetPowerSaving(ctx echo.Context, transactionId models.TransactionId, params models.GetPowerSavingParams) error {
	log := logger.Get()

	transactionIDStr := transactionId.String()
	log.Info("Get power saving request", zap.String("transactionId", transactionIDStr))

	// Retrieve transaction from database
	transaction, err := h.database.GetTransaction(ctx.Request().Context(), transactionIDStr)
	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err), zap.String("transactionId", transactionIDStr))
		return ctx.JSON(http.StatusNotFound, models.ErrorInfo{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "transaction not found",
		})
	}

	// Build device status array
	activationStatus := make([]models.DeviceStatus, 0, len(transaction.Devices))

	for _, txDevice := range transaction.Devices {
		// Determine device status based on completion state and results
		var status models.DeviceStatusStatus

		// If END action exists, use END status (transaction with END has completed or END is in progress)
		if txDevice.EndAction != nil {
			switch txDevice.EndAction.Status {
			case "success":
				status = models.Success
			case "failed":
				status = models.Failed
			case "in-progress":
				status = models.InProgress
			case "pending":
				status = models.Pending
			default:
				status = models.Failed
			}
		} else if txDevice.StartAction != nil {
			switch txDevice.StartAction.Status {
			case "success":
				if transaction.EndAt != nil {
					// END is scheduled but hasn't run yet
					status = models.InProgress
				} else {
					// No END scheduled, START success = complete success
					status = models.Success
				}
			case "failed":
				// START failed = device failed (END won't help)
				status = models.Failed
			case "in-progress":
				status = models.InProgress
			case "pending":
				status = models.Pending
			default:
				status = models.Failed
			}
		} else {
			// No action status yet - default to pending
			status = models.Pending
		}

		activationStatus = append(activationStatus, models.DeviceStatus{
			Device: &txDevice.Device,
			Status: &status,
		})
	}

	// Build response
	response := models.PowerSavingResponse{
		ActivationStatus: &activationStatus,
		TransactionId:    &transactionIDStr,
	}

	return ctx.JSON(http.StatusOK, response)
}
