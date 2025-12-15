# Configuration Reference

The application is configured via environment variables, which are mapped to Helm values.

## Environment Variables

### Common
| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Logging format (`json`, `console`) | `production` (json) |

### API Service
| Variable | Description | Default |
|----------|-------------|---------|
| `API_ADDRESS` | HTTP listen address | `0.0.0.0:8080` |
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `iot` |

### Scheduler Service
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `iot` |
| `RETENTION_PERIOD` | Duration to keep completed transactions | `168h` (7 days) |
| `RETENTION_CLEANUP_INTERVAL` | Frequency of cleanup job | `1h` |

### Worker Service
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `iot` |
| `EASYAPI_BASE_URL` | URL of the 3GPP NEF API | `""` (Dummy Mode) |
| `POWERSAVING_MAX_LATENCY` | Value to set when enabling power saving | `1` |
| `POWERSAVING_MAX_RESPONSE_TIME` | Value to set when enabling power saving | `1` |

### Notifier Service
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | MongoDB database name | `iot` |
| `HTTP_INSECURE_SKIP_VERIFY` | Skip TLS certificate verification for internal cluster services | `false` |

## Helm Values (`values.yaml`)

```yaml
image:
  repository: ghcr.io/your_repository/camara-iot-api
  tag: latest
  pullPolicy: Always

# Services configuration
services:
  api:
    name: iot-api
    minScale: 1
    maxScale: 5
  scheduler:
    name: iot-scheduler
    minScale: 1
    maxScale: 1
  worker:
    name: iot-worker
    minScale: 1
    maxScale: 10
  notifier:
    name: iot-notifier
    minScale: 1
    maxScale: 3
    skipTlsVerify: false  # Set to true for dev/testing with self-signed certs

database:
  uri: "mongodb://root:CHANGEME@mongo-mongodb.data.svc.cluster.local:27017"
  name: iot

easyAPI:
  baseUrl: "" # Set to external URL for production

powerSaving:
  maxLatency: "2"
  maxResponseTime: "2"

retention:
  period: "24h"
  cleanupInterval: "1h"

# Knative Broker Configuration
knative:
  broker:
    name: iot-broker
    class: RabbitMQBroker
    delivery:
      retry: 3
      backoffPolicy: exponential
      backoffDelay: PT1S
  namespace: demo
  triggers:
    parallelism: 3
  rabbitmq:
    brokerConfig:
      name: iot-broker-config
      queueType: quorum
      cluster:
        name: my-rabbit
```
