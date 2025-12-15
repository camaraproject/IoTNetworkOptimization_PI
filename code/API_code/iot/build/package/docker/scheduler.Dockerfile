# syntax=docker/dockerfile:1

# -------- builder --------
FROM golang:1.24-alpine AS builder
WORKDIR /src

# Optional: CA certs for copying to runtime image
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Cache modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/scheduler

# -------- runtime --------
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /out/app /app/app
# copy CA certs for TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
