FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/user-service .

# Create a minimal image
FROM alpine:latest

# Add necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Copy the binary from the builder stage
COPY --from=builder /app/user-service /usr/local/bin/user-service

# Create a non-root user
RUN addgroup -S app && adduser -S app -G app
USER app

# Set the entry point
ENTRYPOINT ["user-service"]

# Expose the port
EXPOSE 8001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O- http://localhost:8001/health || exit 1