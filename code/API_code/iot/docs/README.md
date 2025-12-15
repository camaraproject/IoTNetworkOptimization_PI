# Documentation Overview

This directory contains the documentation for the **IoT Network Optimization API** Provider Implementation, part of the CAMARA project.

The IoT Network Optimization API enables API consumers to manage network-level power-saving features for IoT device fleets. The system uses an event-driven microservices architecture with Knative Eventing and MongoDB, and integrates with 3GPP NEF (via EasyAPI abstraction) to configure network parameters.

## Documentation Index

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture, components, event flows, and database schema |
| [CONFIGURATION.md](CONFIGURATION.md) | Environment variables and Helm configuration reference |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Step-by-step deployment guide for Kubernetes with Helm |
| [REQUIREMENTS.md](REQUIREMENTS.md) | Infrastructure and development environment requirements |
| [MOCKED_COMPONENTS.md](MOCKED_COMPONENTS.md) | Available mocking mechanisms for development and testing |

## Quick Links

- **API Specification**: [`api/iot-network-optimization.yaml`](../api/iot-network-optimization.yaml)
- **CAMARA Project**: [IoTNetworkOptimization](https://github.com/camaraproject/IoTNetworkOptimization)
- **Main README**: [`../README.md`](../README.md)
