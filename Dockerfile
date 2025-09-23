# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy source code
COPY . .

# Download dependencies and build the application
RUN GOPROXY=direct go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/proxy ./cmd/proxy

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