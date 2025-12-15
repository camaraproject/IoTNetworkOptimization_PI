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
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/event"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

const (
	defaultWorkerCount     = 10
	defaultChannelSize     = 100
	defaultRetentionPeriod = 168 * time.Hour // 7 days
	defaultCleanupInterval = 1 * time.Hour
)

// scheduleAction represents a scheduled action with its subscription details
type scheduleAction struct {
	TransactionID       string
	Action              string
	SubscriptionRequest models.SubscriptionRequest
}

// Scheduler handles scheduling and firing of device actuation requests.
type Scheduler struct {
	db              database.Interface
	sender          event.Sender
	receiver        event.Receiver
	fireChan        chan scheduleAction // buffered channel for schedule actions to fire
	workerCount     int
	mu              sync.RWMutex
	timers          map[string]*time.Timer // track active timers for cleanup
	stopCh          chan struct{}
	wg              sync.WaitGroup
	retentionPeriod time.Duration
	cleanupInterval time.Duration
}

// Handler implements receiver.Handler interface for CloudEvents.
type Handler struct {
	scheduler *Scheduler
}

func (h *Handler) Handle(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, error) {
	// Route to appropriate handler based on event type
	switch e.Type() {
	case string(event.EventTypeScheduleRequested):
		return nil, h.scheduler.handleScheduleRequested(ctx, e)
	case string(event.EventTypeAllDevicesCompleted):
		return nil, h.scheduler.handleAllDevicesCompleted(ctx, e)
	default:
		return nil, fmt.Errorf("unknown event type: %s", e.Type())
	}
}

// Config holds scheduler configuration.
type Config struct {
	WorkerCount     int
	ChannelSize     int
	RetentionPeriod time.Duration
	CleanupInterval time.Duration
}

// New creates a new Scheduler instance.
func New(db database.Interface, sender event.Sender, receiver event.Receiver, cfg *Config) *Scheduler {
	if cfg == nil {
		cfg = &Config{
			WorkerCount:     defaultWorkerCount,
			ChannelSize:     defaultChannelSize,
			RetentionPeriod: defaultRetentionPeriod,
			CleanupInterval: defaultCleanupInterval,
		}
	}

	// Apply defaults for zero values
	if cfg.RetentionPeriod == 0 {
		cfg.RetentionPeriod = defaultRetentionPeriod
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = defaultCleanupInterval
	}

	return &Scheduler{
		db:              db,
		sender:          sender,
		receiver:        receiver,
		fireChan:        make(chan scheduleAction, cfg.ChannelSize),
		workerCount:     cfg.WorkerCount,
		timers:          make(map[string]*time.Timer),
		stopCh:          make(chan struct{}),
		retentionPeriod: cfg.RetentionPeriod,
		cleanupInterval: cfg.CleanupInterval,
	}
}

// Start begins processing schedule.requested events and starts the worker pool.
func (s *Scheduler) Start(ctx context.Context) error {
	log := logger.Get()
	log.Info("Starting scheduler",
		zap.Int("workers", s.workerCount),
		zap.Duration("retentionPeriod", s.retentionPeriod),
		zap.Duration("cleanupInterval", s.cleanupInterval))

	// Load pending schedules from database on startup
	if err := s.loadPendingSchedules(ctx); err != nil {
		log.Error("Failed to load pending schedules on startup", zap.Error(err))
		// Continue anyway - new schedules will still work
	}

	// Start worker pool
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i)
	}

	// Start cleanup goroutine for old transactions
	s.wg.Add(1)
	go s.cleanupWorker(ctx)

	// Start event receiver with handler
	handler := &Handler{scheduler: s}
	return s.receiver.Start(handler)
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop(ctx context.Context) error {
	log := logger.Get()
	log.Info("Stopping scheduler")

	close(s.stopCh)

	// Cancel all active timers
	s.mu.Lock()
	for scheduleID, timer := range s.timers {
		timer.Stop()
		delete(s.timers, scheduleID)
	}
	s.mu.Unlock()

	// Close fire channel and wait for workers
	close(s.fireChan)
	s.wg.Wait()

	log.Info("Scheduler stopped")
	return nil
}

// cleanupWorker periodically removes old completed/failed transactions.
func (s *Scheduler) cleanupWorker(ctx context.Context) {
	defer s.wg.Done()
	log := logger.Get()
	log.Info("Starting cleanup worker",
		zap.Duration("interval", s.cleanupInterval),
		zap.Duration("retentionPeriod", s.retentionPeriod))

	// Run initial cleanup (use background context as main ctx may be cancelled during shutdown)
	s.runCleanup()

	// Create ticker for periodic cleanup
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			log.Debug("Cleanup worker stopping")
			return
		case <-ticker.C:
			s.runCleanup()
		}
	}
}

// runCleanup performs the actual cleanup of old transactions.
func (s *Scheduler) runCleanup() {
	log := logger.Get()

	// Use background context with timeout for cleanup operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cutoffTime := time.Now().Add(-s.retentionPeriod)
	log.Debug("Running transaction cleanup",
		zap.Time("cutoffTime", cutoffTime),
		zap.Duration("retentionPeriod", s.retentionPeriod))

	deleted, err := s.db.DeleteOldTransactions(ctx, cutoffTime)
	if err != nil {
		log.Error("Failed to delete old transactions", zap.Error(err))
		return
	}

	if deleted > 0 {
		log.Info("Cleanup completed", zap.Int64("deletedTransactions", deleted))
	} else {
		log.Debug("Cleanup completed, no old transactions to delete")
	}
}

// handleScheduleRequested processes incoming schedule.requested events
func (s *Scheduler) handleScheduleRequested(ctx context.Context, e cloudevents.Event) error {
	log := logger.Get().With(zap.String("eventId", e.ID()), zap.String("eventType", e.Type()))

	var data event.ScheduleRequestedData
	if err := json.Unmarshal(e.Data(), &data); err != nil {
		log.Error("Failed to unmarshal schedule data", zap.Error(err))
		return fmt.Errorf("unmarshal schedule data: %w", err)
	}

	log.Info("Received schedule request",
		zap.String("transactionId", data.Payload.TransactionID),
		zap.Time("startAt", data.StartAt),
		zap.Timep("endAt", data.EndAt))

	// Create Transaction entity with embedded devices
	devices := make([]*database.TransactionDevice, 0, len(data.Payload.Devices))
	for _, device := range data.Payload.Devices {
		if device.NetworkAccessIdentifier == nil {
			log.Error("Device missing networkAccessIdentifier, skipping")
			continue
		}
		deviceID := string(*device.NetworkAccessIdentifier)
		devices = append(devices, &database.TransactionDevice{
			DeviceID: deviceID,
			Device:   device,
			StartAction: &database.DeviceActionStatus{
				Status:    "pending",
				Timestamp: time.Now(),
			},
		})
	}

	transaction := &database.Transaction{
		TransactionID:       data.Payload.TransactionID,
		StartAt:             data.StartAt,
		EndAt:               data.EndAt,
		Enabled:             data.Payload.Enabled,
		SubscriptionRequest: data.Payload.SubscriptionRequest,
		Status:              database.StatusPending,
		Devices:             devices,
	}

	// Create transaction in MongoDB
	if err := s.db.CreateTransaction(ctx, transaction); err != nil {
		log.Error("Failed to create transaction", zap.Error(err), zap.String("transactionId", data.Payload.TransactionID))

		// Send error notification to consumer
		s.sendErrorNotification(ctx, data.Payload.TransactionID, event.ActionStart, "INTERNAL_ERROR", "Failed to create transaction in database", data.Payload.SubscriptionRequest)

		return fmt.Errorf("create transaction: %w", err)
	}

	// Calculate delay until start
	startDelay := time.Until(data.StartAt)
	if startDelay < 0 {
		startDelay = 0 // execute immediately if in the past
	}

	log.Debug("Arming start timer",
		zap.String("transactionId", data.Payload.TransactionID),
		zap.Duration("startDelay", startDelay))

	// Arm time.AfterFunc timer for start
	startTimer := time.AfterFunc(startDelay, func() {
		s.enqueueScheduleAction(data.Payload.TransactionID, event.ActionStart, data.Payload.SubscriptionRequest)
	})

	// Track start timer for cleanup
	startTimerKey := data.Payload.TransactionID + "-start"
	s.mu.Lock()
	s.timers[startTimerKey] = startTimer
	s.mu.Unlock()

	// Note: END timer will be armed after START action completes
	// This is handled in handleAllDevicesCompleted

	return nil
}

// handleAllDevicesCompleted processes all-devices.completed events to arm END timer after START completes
func (s *Scheduler) handleAllDevicesCompleted(ctx context.Context, e cloudevents.Event) error {
	log := logger.Get().With(zap.String("eventId", e.ID()), zap.String("eventType", e.Type()))

	var data event.AllDevicesCompletedData
	if err := json.Unmarshal(e.Data(), &data); err != nil {
		log.Error("Failed to unmarshal all-devices.completed data", zap.Error(err))
		return fmt.Errorf("unmarshal data: %w", err)
	}

	// Check only if START action has completed
	if data.Action != event.ActionStart {
		log.Debug("Ignoring non-START action completion", zap.String("action", data.Action))
		return nil
	}

	log.Debug("START action completed, checking if END timer should be armed",
		zap.String("transactionId", data.TransactionID))

	// Get transaction to check if endAt is defined
	transaction, err := s.db.GetTransaction(ctx, data.TransactionID)
	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err))
		return fmt.Errorf("get transaction: %w", err)
	}

	// If no endAt defined, nothing to do
	if transaction.EndAt == nil {
		log.Debug("No END time defined, skipping END timer",
			zap.String("transactionId", data.TransactionID))
		return nil
	}

	// Calculate delay from NOW until endAt
	endDelay := time.Until(*transaction.EndAt)
	if endDelay < 0 {
		endDelay = 0 // execute immediately if in the past
	}

	log.Debug("Arming END timer after START completion",
		zap.String("transactionId", data.TransactionID),
		zap.Duration("endDelay", endDelay),
		zap.Time("endAt", *transaction.EndAt))

	// Arm timer for end (restore)
	endTimer := time.AfterFunc(endDelay, func() {
		s.enqueueScheduleAction(data.TransactionID, event.ActionEnd, transaction.SubscriptionRequest)
	})

	// Track end timer for cleanup
	endTimerKey := data.TransactionID + "-end"
	s.mu.Lock()
	s.timers[endTimerKey] = endTimer
	s.mu.Unlock()

	return nil
}

// loadPendingSchedules loads all pending transactions from database and arms their timers
func (s *Scheduler) loadPendingSchedules(ctx context.Context) error {
	log := logger.Get()
	log.Info("Loading pending schedules from database")

	transactions, err := s.db.GetPendingTransactions(ctx)
	if err != nil {
		return fmt.Errorf("get pending transactions: %w", err)
	}

	if len(transactions) == 0 {
		log.Info("No pending schedules to load")
		return nil
	}

	log.Info("Restoring pending schedules", zap.Int("count", len(transactions)))

	for _, tx := range transactions {
		// Check if START action needs to be scheduled
		if !tx.StartActionCompleted {
			startDelay := time.Until(tx.StartAt)
			if startDelay < 0 {
				startDelay = 0 // Execute immediately if in the past
			}

			log.Debug("Arming START timer for pending transaction",
				zap.String("transactionId", tx.TransactionID),
				zap.Duration("delay", startDelay))

			startTimer := time.AfterFunc(startDelay, func() {
				s.enqueueScheduleAction(tx.TransactionID, event.ActionStart, tx.SubscriptionRequest)
			})

			startTimerKey := tx.TransactionID + "-start"
			s.mu.Lock()
			s.timers[startTimerKey] = startTimer
			s.mu.Unlock()
		}

		// Check if END action needs to be scheduled
		// END is only scheduled if START is already completed and END time is defined
		if tx.EndAt != nil && tx.StartActionCompleted && !tx.EndActionCompleted {
			endDelay := time.Until(*tx.EndAt)
			if endDelay < 0 {
				endDelay = 0 // Execute immediately if in the past
			}

			log.Debug("Arming END timer for pending transaction",
				zap.String("transactionId", tx.TransactionID),
				zap.Duration("delay", endDelay))

			endTimer := time.AfterFunc(endDelay, func() {
				s.enqueueScheduleAction(tx.TransactionID, event.ActionEnd, tx.SubscriptionRequest)
			})

			endTimerKey := tx.TransactionID + "-end"
			s.mu.Lock()
			s.timers[endTimerKey] = endTimer
			s.mu.Unlock()
		}
	}

	log.Info("Pending schedules restored successfully", zap.Int("loaded", len(transactions)))
	return nil
}

// enqueueScheduleAction helper to enqueue a schedule action (start or end) to the fire channel
func (s *Scheduler) enqueueScheduleAction(transactionID string, action string, subscriptionRequest models.SubscriptionRequest) {
	log := logger.Get().With(
		zap.String("transactionId", transactionID),
		zap.String("action", action))

	// Create schedule action with subscription details
	schedAction := scheduleAction{
		TransactionID:       transactionID,
		Action:              action,
		SubscriptionRequest: subscriptionRequest,
	}

	// Enqueue to fire channel (non-blocking callback)
	select {
	case s.fireChan <- schedAction:
		log.Debug("Schedule action enqueued for firing")
	case <-s.stopCh:
		log.Debug("Scheduler stopped, skipping fire")
	default:
		log.Warn("Fire channel full, schedule may be delayed")
		// Try again with blocking send
		select {
		case s.fireChan <- schedAction:
		case <-s.stopCh:
		}
	}

	// Remove timer from map
	timerKey := transactionID + "-" + action
	s.mu.Lock()
	delete(s.timers, timerKey)
	s.mu.Unlock()
}

// worker processes fired schedules from the channel
func (s *Scheduler) worker(ctx context.Context, id int) {
	defer s.wg.Done()
	log := logger.Get().With(zap.Int("workerId", id))
	log.Debug("Scheduler worker started")

	for {
		select {
		case <-ctx.Done():
			log.Debug("Worker stopping due to context cancellation")
			return
		case <-s.stopCh:
			log.Debug("Worker stopping")
			return
		case schedAction, ok := <-s.fireChan:
			if !ok {
				log.Debug("Worker stopping, channel closed")
				return
			}

			if err := s.fireSchedule(ctx, schedAction); err != nil {
				log.Error("Failed to fire schedule", zap.Error(err), zap.String("scheduleId", schedAction.TransactionID))
			}
		}
	}
}

// fireSchedule atomically claims and publishes individual device.actuation.request events
func (s *Scheduler) fireSchedule(ctx context.Context, schedAction scheduleAction) error {
	log := logger.Get().With(
		zap.String("transactionId", schedAction.TransactionID),
		zap.String("action", schedAction.Action))

	// Atomically claim transaction for this specific action
	claimed, err := s.db.ClaimTransaction(ctx, schedAction.TransactionID, schedAction.Action)
	if err != nil {
		log.Error("Failed to claim transaction", zap.Error(err))

		// Send error notification to consumer using cached SubscriptionRequest
		s.sendErrorNotification(ctx, schedAction.TransactionID, schedAction.Action, "INTERNAL_ERROR", "Failed to claim transaction in database", schedAction.SubscriptionRequest)

		return fmt.Errorf("claim transaction: %w", err)
	}

	if !claimed {
		log.Debug("Transaction action already completed or claimed, skipping")
		return nil
	}

	log.Debug("Transaction claimed, publishing actuation requests")

	// Retrieve full transaction data
	transaction, err := s.db.GetTransaction(ctx, schedAction.TransactionID)
	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err))
		_ = s.db.MarkTransactionFailed(ctx, schedAction.TransactionID, "failed to retrieve transaction data")

		// Send error notification to consumer using cached SubscriptionRequest
		s.sendErrorNotification(ctx, schedAction.TransactionID, schedAction.Action, "INTERNAL_ERROR", "Failed to retrieve transaction data from database", schedAction.SubscriptionRequest)

		return fmt.Errorf("get transaction: %w", err)
	}

	// Determine enabled value: for start use transaction.Enabled, for end use !transaction.Enabled
	enabledValue := transaction.Enabled
	if schedAction.Action == event.ActionEnd {
		enabledValue = !transaction.Enabled // Invert for end action
	}

	// Publish individual device.actuation.request event for each device
	log.Debug("Publishing device actuation requests",
		zap.Int("deviceCount", len(transaction.Devices)),
		zap.Bool("enabled", enabledValue))
	for i, txDevice := range transaction.Devices {
		actuationData := event.DeviceActuationRequestData{
			Device:              txDevice.Device,
			Enabled:             enabledValue,
			TransactionID:       transaction.TransactionID,
			Action:              schedAction.Action,
			SubscriptionRequest: schedAction.SubscriptionRequest,
		}

		// Use unique event ID for each device
		eventID := fmt.Sprintf("%s-%s-device-%d", transaction.TransactionID, schedAction.Action, i)
		err = s.sender.Send(
			ctx,
			eventID,
			event.EventTypeDeviceActuationRequest,
			event.SourceiotScheduler,
			actuationData,
		)
		if err != nil {
			log.Error("Failed to publish actuation request for device",
				zap.Error(err),
				zap.Int("deviceIndex", i))
			// Continue with other devices even if one fails
		}
	}

	log.Debug("Device actuation requests published successfully")
	return nil
}

// sendErrorNotification sends a CloudEventError to the consumer's notification sink
func (s *Scheduler) sendErrorNotification(ctx context.Context, transactionID string, action string, errorCode string, errorMessage string, subscriptionRequest models.SubscriptionRequest) {
	log := logger.Get().With(
		zap.String("transactionId", transactionID),
		zap.String("errorCode", errorCode))

	// Check if subscription request has a sink
	if subscriptionRequest.Sink == "" {
		log.Warn("No notification sink configured, skipping error notification")
		return
	}

	// Prepare error notification data
	errorData := event.ErrorNotificationData{
		TransactionID:       transactionID,
		Status:              500,
		Code:                errorCode,
		Message:             errorMessage,
		Action:              action,
		SubscriptionRequest: subscriptionRequest,
	}

	// Send error event to notification sink (via notifier service)
	eventID := fmt.Sprintf("%s-%s-error", transactionID, action)
	if err := s.sender.Send(ctx, eventID, event.EventTypePowerSavingError, event.SourceiotScheduler, errorData); err != nil {
		log.Error("Failed to send error notification event", zap.Error(err))
	} else {
		log.Info("Error notification event sent successfully")
	}
}
