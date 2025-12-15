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
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/worker"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/easyapi"
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

	log.Info("Starting IoT Actuation Worker Service")

	// Initialize database
	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Info("Database connected")

	// Initialize device client
	var deviceClient easyapi.Client
	if conf.EasyAPI.BaseURL != "" {
		deviceClient = easyapi.New(conf.EasyAPI.BaseURL)
		log.Info("Device client initialized (EasyAPI mode)",
			zap.String("baseURL", conf.EasyAPI.BaseURL))
	} else {
		deviceClient = easyapi.NewDummy()
		log.Info("Device client initialized (DUMMY mode - no real API calls)")
	}

	// Initialize event sender (requires K_SINK environment variable)
	sender, err := event.NewSender()
	if err != nil {
		return fmt.Errorf("failed to create event sender: %w", err)
	}
	log.Info("Event sender initialized")

	// Initialize event receiver for device.actuation.request events
	receiver, err := event.NewReceiver(conf.API)
	if err != nil {
		return fmt.Errorf("failed to create event receiver: %w", err)
	}
	log.Info("Event receiver initialized", zap.String("address", conf.API.Address))

	// Create actuation worker
	actuationWorker := worker.New(db, deviceClient, sender, receiver, conf.PowerSaving)
	log.Info("Power saving configuration loaded",
		zap.String("maxLatency", conf.PowerSaving.MaxLatency),
		zap.String("maxResponseTime", conf.PowerSaving.MaxResponseTime))

	// Setup graceful shutdown
	_ = context.Background() // context not currently used but available for future use

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start worker in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := actuationWorker.Start(); err != nil {
			errChan <- fmt.Errorf("worker error: %w", err)
		}
	}()

	log.Info("Actuation worker service is running")

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Info("Received shutdown signal")
	case err := <-errChan:
		log.Error("Worker error", zap.Error(err))
		return err
	}

	log.Info("Worker service stopped")
	return nil
}
