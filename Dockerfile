# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN apk add --no-cache wget

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o alertbridge ./cmd/alertbridge

# Final stage
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/alertbridge .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./alertbridge"] 