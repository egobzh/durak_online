# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /durak-server ./cmd/server/main.go

# Run stage
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates and curl for healthchecks
RUN apk --no-cache add ca-certificates curl tzdata

COPY --from=builder /durak-server /app/durak-server

EXPOSE 8080

ENTRYPOINT ["/app/durak-server"]
