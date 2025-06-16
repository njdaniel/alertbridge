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
FROM gcr.io/distroless/static-debian11@sha256:e6d589f36c6c7d9a14df69da026b446ac03c0d2027bfca82981b6a1256c2019c

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/alertbridge .
# (Line removed as it is redundant)

# Expose port
EXPOSE 3000

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=5s CMD ["/usr/bin/wget","-qO-","http://localhost:3000/healthz"]

# Run the application
CMD ["./alertbridge"] 