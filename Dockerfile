# Builder stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/api ./cmd/api

# Final stage: distroless static nonroot
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/api /app/api

# Run as nonroot
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/api"]
