# Requirements

To build, deploy, and run the CAMARA IoT Network Optimization API, you need the following infrastructure and tools.

## Development Environment

*   **Go**: Version 1.24.1 or higher.
*   **Docker**: For building container images.
*   **Make** (Optional): For running build scripts if provided.
*   **OAPI-Codegen**: For regenerating API code from OpenAPI specs (`go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`).

## Infrastructure

The system is designed to run on Kubernetes with Knative Eventing.

### Kubernetes Cluster
*   A standard Kubernetes cluster (v1.24+).
*   Local options: [Kind](https://kind.sigs.k8s.io/), [Minikube](https://minikube.sigs.k8s.io/), or Docker Desktop.

### Knative Eventing & RabbitMQ
*   **Knative Eventing** must be installed on the cluster.
*   **RabbitMQ Broker**: The Helm chart is configured to use the `RabbitMQBroker` class. You must have the **Knative RabbitMQ Controller** installed.
*   **RabbitMQ Cluster**: A RabbitMQ cluster must be available in the Kubernetes cluster (e.g., via the RabbitMQ Cluster Operator). The Helm chart expects to link the broker to this cluster.

### Database
*   **MongoDB**: Version 5.0+.
*   The services require a connection string (URI) to a MongoDB instance.
*   For development, a simple MongoDB pod/service in the cluster is sufficient.
*   For production, a managed MongoDB service or a high-availability replica set is recommended.

### 3GPP Network Exposure Function (NEF) / EasyAPI
*   The system expects to communicate with a 3GPP-compliant API (referred to as "EasyAPI" in this codebase) to configure device network parameters (`/nudm-sdm` and `/nudm-pp`).
*   If a real NEF is not available, the system includes a **Dummy Client** mode and a **Sink Receiver** tool for testing.

## Network Requirements

*   **Ingress Controller**: To expose the API service to the outside world (e.g., Nginx Ingress, Istio, or Kourier for Knative).
*   **Egress**: The Worker service needs network access to the 3GPP NEF/EasyAPI.
*   **Internal DNS**: Services must be able to resolve the MongoDB service and the Knative Broker.
