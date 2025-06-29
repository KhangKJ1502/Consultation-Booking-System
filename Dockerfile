# Build stage
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application - Fixed path to match actual directory structure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and netcat for health checks
RUN apk --no-cache add ca-certificates netcat-openbsd

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy environment files
COPY --from=builder /app/environment ./environment

# Create a script to wait for dependencies
RUN echo '#!/bin/sh' > wait-for-it.sh && \
    echo 'until nc -z kafka 9092; do' >> wait-for-it.sh && \
    echo '  echo "Waiting for Kafka..."' >> wait-for-it.sh && \
    echo '  sleep 2' >> wait-for-it.sh && \
    echo 'done' >> wait-for-it.sh && \
    echo 'echo "Kafka is ready!"' >> wait-for-it.sh && \
    echo 'exec "$@"' >> wait-for-it.sh && \
    chmod +x wait-for-it.sh

EXPOSE 8080

# Use the wait script as entrypoint
ENTRYPOINT ["./wait-for-it.sh"]
CMD ["./main"]