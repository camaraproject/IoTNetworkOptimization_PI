FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the notifier binary
RUN CGO_ENABLED=0 GOOS=linux go build -o notifier ./cmd/notifier

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/notifier .

EXPOSE 8080

CMD ["./notifier"]
