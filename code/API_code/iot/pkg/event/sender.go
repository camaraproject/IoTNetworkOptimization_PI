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
	"context"
	"fmt"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/logger"
)

type Sender interface {
	Send(ctx context.Context, requestID string, eventType EventType, source Source, data any, opts ...Option) error
}

type sender struct {
	client cloudevents.Client
}

type Option func(*cloudevents.Event)

func WithSubject(sub string) Option {
	return func(e *cloudevents.Event) { e.SetSubject(sub) }
}

// NewSender creates a Sender using the K_SINK environment variable.
func NewSender() (Sender, error) {
	target := os.Getenv("K_SINK")
	if target == "" {
		return nil, fmt.Errorf("missing broker URL: set via SinkBinding or K_SINK env var")
	}
	c, err := cloudevents.NewClientHTTP(cloudevents.WithTarget(target))
	if err != nil {
		return nil, err
	}
	return &sender{client: c}, nil
}

func (s *sender) Send(ctx context.Context, id string, eventType EventType, source Source, data any, opts ...Option) (err error) {
	log := logger.Get()
	log.With(
		zap.String("event-id", id),
		zap.String("event-type", eventType.String()),
		zap.String("source", source.String()),
		zap.Any("data", data),
	).Debug("Sending cloud event")
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in CloudEvents sender: %v", r)
		}
	}()

	e, err := Event(id, eventType, source, data, opts...)
	if err != nil {
		return nil
	}
	if res := s.client.Send(ctx, *e); cloudevents.IsUndelivered(res) {
		log.With(zap.Error(res)).Error("Send cloud event failed")
		return res
	}
	log.Debug("Cloud event sent")
	return nil
}
