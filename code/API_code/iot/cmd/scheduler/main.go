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
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/scheduler"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/event"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize logger and configuration
	conf := config.GetConf()
	log := logger.Get()

	log.Info("Starting IoT Scheduler Service")

	// Initialize database
	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Info("Database connected")

	// Initialize event sender (requires K_SINK environment variable)
	sender, err := event.NewSender()
	if err != nil {
		return fmt.Errorf("failed to create event sender: %w", err)
	}
	log.Info("Event sender initialized")

	// Initialize event receiver for schedule.requested events
	receiver, err := event.NewReceiver(conf.API)
	if err != nil {
		return fmt.Errorf("failed to create event receiver: %w", err)
	}
	log.Info("Event receiver initialized", zap.String("address", conf.API.Address))

	// Parse retention configuration
	retentionPeriod, err := time.ParseDuration(conf.Retention.Period)
	if err != nil {
		log.Warn("Invalid retention period, using default 168h",
			zap.String("configured", conf.Retention.Period),
			zap.Error(err))
		retentionPeriod = 168 * time.Hour
	}

	cleanupInterval, err := time.ParseDuration(conf.Retention.CleanupInterval)
	if err != nil {
		log.Warn("Invalid cleanup interval, using default 1h",
			zap.String("configured", conf.Retention.CleanupInterval),
			zap.Error(err))
		cleanupInterval = 1 * time.Hour
	}

	// Create scheduler with custom config
	schedulerCfg := &scheduler.Config{
		WorkerCount:     10,  // number of workers processing fired schedules
		ChannelSize:     100, // buffered channel size
		RetentionPeriod: retentionPeriod,
		CleanupInterval: cleanupInterval,
	}

	log.Info("Retention configuration",
		zap.Duration("retentionPeriod", retentionPeriod),
		zap.Duration("cleanupInterval", cleanupInterval))

	sched := scheduler.New(db, sender, receiver, schedulerCfg)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start scheduler in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := sched.Start(ctx); err != nil {
			errChan <- fmt.Errorf("scheduler error: %w", err)
		}
	}()

	log.Info("Scheduler service is running")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Info("Received shutdown signal")
	case err := <-errChan:
		log.Error("Scheduler error", zap.Error(err))
		return err
	}

	// Graceful shutdown
	log.Info("Shutting down scheduler")
	if err := sched.Stop(ctx); err != nil {
		log.Error("Error during shutdown", zap.Error(err))
		return err
	}

	log.Info("Scheduler service stopped")
	return nil
}
