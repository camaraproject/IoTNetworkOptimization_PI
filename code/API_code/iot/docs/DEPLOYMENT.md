# Deployment Guide

This guide explains how to deploy the CAMARA IoT Network Optimization API to a Kubernetes cluster using Helm.

## Prerequisites

1.  A Kubernetes cluster with **Knative Eventing** installed.
2.  **RabbitMQ Cluster** and **Knative RabbitMQ Controller** (the Helm chart uses `RabbitMQBroker`).
3.  **Helm** 3.0+ installed locally.
4.  **MongoDB** running in the cluster (or accessible from it).

## Build Images

Before deploying, you need to build the Docker images for the microservices.

```bash
# Build API
docker build -f build/package/docker/api.Dockerfile -t your-registry/iot-api:latest .

# Build Scheduler
docker build -f build/package/docker/scheduler.Dockerfile -t your-registry/iot-scheduler:latest .

# Build Worker
docker build -f build/package/docker/worker.Dockerfile -t your-registry/iot-worker:latest .

# Build Notifier
docker build -f build/package/docker/notifier.Dockerfile -t your-registry/iot-notifier:latest .

# Push images
docker push your-registry/iot-api:latest
# ... repeat for others
```

## Helm Deployment

The project includes a Helm chart in `deploy/helm/api`.

### 1. Configure Values

Create a `my-values.yaml` file to override the defaults.

```yaml
# my-values.yaml

# Image configuration
image:
  repository: your-registry
  tag: latest
  pullPolicy: Always

# Database Connection
database:
  uri: "mongodb://username:password@mongo-service:27017"
  name: "iot_db"

# 3GPP API Configuration
easyAPI:
  baseUrl: "http://nef-service.telecom.com" # Leave empty to use internal Dummy client

# Power Saving Parameters
powerSaving:
  maxLatency: "100"
  maxResponseTime: "200"

# Data Retention
retention:
  period: "168h" # 7 days
  cleanupInterval: "1h"
```

### 2. Install Chart

```bash
helm install iot-api ./deploy/helm/api \
  --namespace camara-iot \
  --create-namespace \
  -f my-values.yaml
```

### 3. Verify Deployment

Check that all pods are running:

```bash
kubectl get pods -n camara-iot
```

You should see pods for:
*   `iot-api`
*   `iot-scheduler`
*   `iot-worker`
*   `iot-notifier`

Check that the Knative triggers are ready:

```bash
kubectl get triggers -n camara-iot
```

## Local Development (Kind/Minikube)

For local testing, you can use the provided `sinkreceiver` to mock the 3GPP API and the callback receiver.

1.  **Install MongoDB**:
    ```bash
    helm install mongo oci://registry-1.docker.io/bitnamicharts/mongodb -n camara-iot
    ```

2.  **Deploy the API with Dummy Mode**:
    Set `easyAPI.baseUrl` to `""` (empty string) or point it to the `sinkreceiver` service if you deploy it.

    ```bash
    helm install iot-api ./deploy/helm/api \
      --namespace camara-iot \
      --set database.uri="mongodb://root:password@mongo-mongodb.camara-iot.svc.cluster.local:27017" \
      --set easyAPI.baseUrl="" 
    ```
