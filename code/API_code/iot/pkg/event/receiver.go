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
	"net"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/internal/receiver"
	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
)

// Receiver starts an HTTP server that delivers incoming CloudEvents to a handler.
type Receiver interface {
	Start(fn receiver.Handler) error
}

type eventReceiver struct {
	client cloudevents.Client
}

// NewReceiver creates a CloudEvents HTTP server bound to the configured address.
func NewReceiver(conf config.API) (Receiver, error) {
	ln, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, err
	}
	protocol, err := cloudevents.NewHTTP(cloudevents.WithListener(ln))
	if err != nil {
		return nil, err
	}
	client, err := cloudevents.NewClient(protocol)
	if err != nil {
		return nil, err
	}

	return &eventReceiver{client: client}, nil
}

// Start runs the server and delivers events to fn until ctx is done.
func (r *eventReceiver) Start(handler receiver.Handler) error {
	return r.client.StartReceiver(context.TODO(), handler.Handle)
}
