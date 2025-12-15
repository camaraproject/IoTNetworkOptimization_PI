# Architecture

This project implements the CAMARA IoT Network Optimization API using an event-driven microservices architecture.

## Overview

The system is composed of several decoupled services that communicate primarily through CloudEvents and a shared MongoDB database.

### Components

1.  **API Service (`cmd/api`)**
    *   **Role**: Entry point for API consumers.
    *   **Responsibilities**:
        *   Validates incoming requests against the OpenAPI specification.
        *   Resolves device identifiers (e.g., converting Phone Number to NAI).
        *   Checks for conflicting transactions.
        *   Creates transaction records in MongoDB with `pending` status.
        *   Publishes `schedule.requested` events to the event broker.
    *   **Tech**: Go, Echo Framework, OAPI-Codegen.

2.  **Scheduler Service (`cmd/scheduler`)**
    *   **Role**: Manages the timing of device actuations.
    *   **Responsibilities**:
        *   Listens for `schedule.requested` events.
        *   Manages in-memory timers for `START` and `END` actions.
        *   Persists schedule state to allow recovery after restarts.
        *   When a timer fires, it atomically claims the transaction action in the DB.
        *   Publishes `device.actuation.request` events for each device in the transaction.
        *   Runs a background cleanup job to remove old completed transactions.
    *   **Tech**: Go, CloudEvents SDK, time.Timer.

3.  **Worker Service (`cmd/worker`)**
    *   **Role**: Executes the actual device configuration changes.
    *   **Responsibilities**:
        *   Listens for `device.actuation.request` events.
        *   Interacts with the 3GPP Network Exposure Function (via the `EasyAPI` interface).
        *   **Start Action**: Backs up the current device configuration and applies the power-saving profile.
        *   **End Action**: Restores the original device configuration.
        *   Updates device status in MongoDB (`in-progress` -> `success`/`failed`).
        *   Detects when all devices in a transaction have completed an action and publishes `all-devices.completed`.
    *   **Tech**: Go, CloudEvents SDK.

4.  **Notifier Service (`cmd/notifier`)**
    *   **Role**: Handles callbacks to the API consumer.
    *   **Responsibilities**:
        *   Listens for `all-devices.completed` events.
        *   Retrieves the full transaction status from MongoDB.
        *   Sends a webhook notification to the `notificationUrl` provided in the initial request.
    *   **Tech**: Go, CloudEvents SDK.

5.  **Sink Receiver (`cmd/sinkreceiver`)**
    *   **Role**: Testing utility.
    *   **Responsibilities**:
        *   Acts as a mock endpoint for the `EasyAPI` (3GPP interface) during local development.
        *   Can be used to verify that the Worker is making the correct calls.

## Knative Eventing

The system relies heavily on Knative Eventing for asynchronous communication. The following table describes the events and their flow through the system.

| Event Type | Source | Producer | Consumer(s) | Description |
| :--- | :--- | :--- | :--- | :--- |
| `it.tim.iot.schedule.requested` | `urn:tim:iot-api` | **API** | **Scheduler** | Sent when a user creates a new power-saving schedule. Contains the transaction ID and schedule details. |
| `it.tim.iot.device.actuation.request` | `urn:tim:iot-scheduler` | **Scheduler** | **Worker** | Sent when a schedule timer fires (Start or End). Contains the transaction ID, action type (`start`/`end`), and the list of devices to actuate. |
| `it.tim.iot.all-devices.completed` | `urn:tim:iot-worker` | **Worker** | **Notifier**, **Scheduler** | Sent when the Worker has finished processing all devices for a specific action. <br>• **Notifier**: Uses this to send the webhook callback.<br>• **Scheduler**: Uses this to arm the "End" timer after the "Start" action completes. |
| `it.tim.iot.notify.error.requested` | `urn:tim:iot-notify` | **Notifier** | - | Sent when a system-level error prevents processing. Contains error details and the affected transaction. |

### Triggers

The following Knative Triggers are defined to route events from the Broker to the services:

*   `schedule-requested-trigger`: Routes `schedule.requested` -> `iot-scheduler`.
*   `device-actuation-trigger`: Routes `device.actuation.request` -> `iot-worker`.
*   `all-devices-completed-notifier-trigger`: Routes `all-devices.completed` -> `iot-notifier`.
*   `all-devices-completed-scheduler-trigger`: Routes `all-devices.completed` -> `iot-scheduler`.

## Data Flow

1.  **Request**: User sends `POST /features/power-saving`.
2.  **Validation**: API validates request and resolves identifiers.
3.  **Persistence**: API creates a Transaction document in MongoDB.
4.  **Event**: API sends `schedule.requested` to Broker.
5.  **Scheduling**: Scheduler receives event, sets timers for Start (and optional End) times.
6.  **Firing**: Timer fires. Scheduler sends `device.actuation.request` (one per device) to Broker.
7.  **Actuation**: Worker receives request.
    *   Calls 3GPP API (EasyAPI) to apply config.
    *   Updates MongoDB device status.
8.  **Completion**: Worker checks if all devices for the transaction are done. If so, sends `all-devices.completed`.
9.  **Notification**: Notifier receives completion event and sends webhook to user.

## Database Schema

The system uses MongoDB with the following primary collections:

### `transactions`
Stores the state of every power-saving request. This is the primary source of truth for the system state.

*   `_id` (String): Unique Transaction ID (UUID).
*   `startAt` (Date): Scheduled start time for the power-saving profile.
*   `endAt` (Date, Optional): Scheduled end time.
*   `enabled` (Boolean): Whether the power saving mode is being enabled or disabled.
*   `subscriptionRequest` (Object): Callback details.
    *   `sink` (String): The webhook URL.
    *   `sinkCredential` (Object): Auth token (if provided).
*   `status` (String): Overall transaction status (`pending`, `processing`, `completed`, `failed`).
*   `createdAt` (Date): Creation timestamp.
*   `updatedAt` (Date): Last update timestamp.
*   `errorMessage` (String, Optional): Error details if the transaction failed.
*   `devices` (Array): List of devices included in this transaction.
    *   `deviceId` (String): Internal device identifier (NAI).
    *   `device` (Object): Original device identifier provided by the user (e.g., `phoneNumber`).
    *   `startAction` (Object): Status of the activation operation.
        *   `status` (String): `in-progress`, `success`, `failed`.
        *   `timestamp` (Date): Time of the last status change.
    *   `endAction` (Object): Status of the deactivation operation.
        *   `status` (String): `in-progress`, `success`, `failed`.
        *   `timestamp` (Date): Time of the last status change.
*   `startActionCompleted` (Boolean): True if the start action has been processed for all devices.
*   `endActionCompleted` (Boolean): True if the end action has been processed for all devices.
*   `startActionNotified` (Boolean): True if the start completion notification has been sent.
*   `endActionNotified` (Boolean): True if the end completion notification has been sent.

### `device_configs`
Stores the original state of devices before power-saving was applied. This allows the system to restore the exact previous configuration when the power-saving period ends.

*   `_id` (String): Device ID (NAI).
*   `ppMaximumLatency` (String): The original latency setting.
*   `ppMaximumResponseTime` (String): The original response time setting.
*   `timestamp` (Date): When this configuration was backed up.
