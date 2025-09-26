# syntax=docker/dockerfile:1

ARG GO_VERSION="1.24-alpine"

FROM golang:${GO_VERSION} AS builder

RUN apk update && apk add --no-cache git bash ca-certificates && update-ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chat-client ./client
RUN chmod +x ./chat-client

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/chat-client /chat-client

ENTRYPOINT ["/chat-client"]