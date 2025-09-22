# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go modules and vendor directory
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy source code
COPY . .

# Build the application with static linking using vendor
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/proxy ./cmd/proxy

# Final stage - use distroless for security and minimal size  
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/bin/proxy /proxy

# Copy configuration from examples
COPY --from=builder /app/examples/config.yaml /config.yaml

# Expose ports
EXPOSE 8080 9090

# Run the proxy
CMD ["/proxy", "-config", "/config.yaml"]