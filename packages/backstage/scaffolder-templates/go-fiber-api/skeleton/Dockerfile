# Multi-stage build for minimal image size
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server .

# Expose port
EXPOSE ${{ values.port }}

# Run
CMD ["./server"]
