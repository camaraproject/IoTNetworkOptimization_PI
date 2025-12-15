# API Code Generation

This folder contains the OpenAPI specification and generated Go code using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Structure

- `iot-network-optimization.yaml`: OpenAPI schema definition.
- `config/`: Configuration files for code generation (client, server, models).
- `client/`, `server/`, `models/`: Generated Go code.
- `generate.go`: Contains `go:generate` directives for code generation.

## How to Generate Code

Use this command from the root of the repo:

```bash
go generate ./...
```

This will execute the generation for client, server, and models using their corresponding configuration files in `config/`.

## Versioning

We use Go 1.24+ `go get -tool` support for installing `oapi-codegen`. To upgrade:

```bash
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@vX.X.X
```

## Auto-generation on Save

If you are using VS Code and the [emeraldwalk.runonsave](https://marketplace.visualstudio.com/items?itemName=emeraldwalk.RunOnSave) extension, types will regenerate automatically on saving the `.yaml` spec file.
