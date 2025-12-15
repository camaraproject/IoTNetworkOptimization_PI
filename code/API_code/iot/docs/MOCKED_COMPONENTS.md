# Mocked components

To facilitate development and testing without requiring access to a live 5G Core Network or 3GPP NEF, this project includes several mocking mechanisms.

## 1. Dummy EasyAPI Client

The **Worker** service includes an internal "Dummy" implementation of the `EasyAPI` client interface.

*   **Activation**: This mode is activated automatically if the `EASYAPI_BASE_URL` environment variable (or `easyAPI.baseUrl` in Helm) is set to an empty string `""`.
*   **Behavior**:
    *   **GetDeviceConfig**: Returns a static configuration (Latency: 100, ResponseTime: 200).
    *   **SetDeviceConfig**: Logs the request and returns success immediately.
*   **Use Case**: Unit testing, local development where no external network is available.

## 2. Sink Receiver (`cmd/sinkreceiver`)

The **Sink Receiver** is a standalone service that can act as an external mock for the 3GPP API.

*   **Role**: It starts an HTTP server that logs all incoming requests.
*   **Usage**:
    1.  Deploy the `sinkreceiver` (chart available in `deploy/helm/sinkreceiver`).
    2.  Configure the main API's `EASYAPI_BASE_URL` to point to the sink receiver service (e.g., `http://iotsinkreceiver:8090`).
*   **Behavior**: It accepts any request and logs the body/headers. This allows you to verify that the Worker is sending the correct HTTP requests (paths, payloads, headers) without needing a real backend.

## 3. Device Identifier Mock Translator

The API needs to translate various device identifiers (PhoneNumber, IPv4, IPv6) into a canonical `NetworkAccessIdentifier` (NAI) for internal processing.

*   **Implementation**: `pkg/deviceidentifier/mock.go`
*   **Logic**:
    *   If an NAI is provided in the request, it is used as-is.
    *   If other identifiers are provided, it generates a deterministic **SHA-256 hash** of the identifiers and creates a fake NAI in the format: `{hash}@generated.nai`.
*   **Purpose**: Allows the system to function consistently with arbitrary input data without needing a real subscriber database (UDM/HSS) to perform lookups.
