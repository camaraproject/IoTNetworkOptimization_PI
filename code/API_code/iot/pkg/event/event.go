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
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	partitionKey string = "partitionkey"
)

func Event(requestID string, eventType EventType, source Source, data any, opts ...Option) (*cloudevents.Event, error) {
	e := cloudevents.NewEvent()
	e.SetID(requestID)
	e.SetExtension(partitionKey, requestID)
	e.SetSource(source.String())
	e.SetType(eventType.String())
	e.SetTime(time.Now())
	if err := e.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(&e)
	}
	return &e, nil
}
