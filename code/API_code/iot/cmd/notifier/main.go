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

	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/database"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/notifier"
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

	log.Info("Starting IoT Notification Worker Service")

	// Initialize database
	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Info("Database connected")

	// Initialize event receiver for all-devices.completed events
	receiver, err := event.NewReceiver(conf.API)
	if err != nil {
		return fmt.Errorf("failed to create event receiver: %w", err)
	}
	log.Info("Event receiver initialized", zap.String("address", conf.API.Address))

	// Create notification worker
	notificationWorker := notifier.New(db, receiver)

	// Setup graceful shutdown
	_ = context.Background() // context not currently used but available for future use

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start worker in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := notificationWorker.Start(); err != nil {
			errChan <- fmt.Errorf("worker error: %w", err)
		}
	}()

	log.Info("Notification worker service is running")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Info("Received shutdown signal")
	case err := <-errChan:
		log.Error("Worker error", zap.Error(err))
		return err
	}

	log.Info("Notification worker service stopped")
	return nil
}
