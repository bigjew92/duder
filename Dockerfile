# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . ./

# Build with optimizations: disable CGO, strip debug symbols
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /duder

# Final stage - minimal Alpine image
FROM alpine:latest

# Add CA certs for HTTPS requests
RUN apk --no-cache add ca-certificates

COPY --from=builder /duder /duder

CMD ["/duder"]