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

type EventType string

const (
	// EventTypeScheduleRequested is sent by the API to create a new schedule.
	EventTypeScheduleRequested EventType = "it.tim.iot.schedule.requested"

	// EventTypeDeviceActuationRequest is sent by the Scheduler to perform device actuation.
	EventTypeDeviceActuationRequest EventType = "it.tim.iot.device.actuation.request"

	// EventTypeAllDevicesCompleted is sent when all devices for a transaction have completed.
	EventTypeAllDevicesCompleted EventType = "it.tim.iot.all-devices.completed"

	// EventTypePowerSavingError is sent when a system-level error prevents processing.
	EventTypePowerSavingError EventType = "it.tim.iot.notify.error.requested"
)

func (s EventType) String() string {
	return string(s)
}

type Source string

const (
	// SourceiotAPI is the CloudEvents source for the API service.
	SourceiotAPI Source = "urn:tim:iot-api"

	// SourceiotScheduler is the CloudEvents source for the Scheduler service.
	SourceiotScheduler Source = "urn:tim:iot-scheduler"

	// SourceiotWorker is the CloudEvents source for the Worker service.
	SourceiotWorker Source = "urn:tim:iot-worker"

	// SourceiotNotify is the CloudEvents source for the Notify service.
	SourceiotNotify Source = "urn:tim:iot-notify"
)

func (s Source) String() string {
	return string(s)
}
