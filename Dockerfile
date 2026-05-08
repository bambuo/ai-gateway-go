# Multi-stage build
FROM golang:1.26-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /gateway ./cmd/gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /gateway /gateway

EXPOSE 8443
ENTRYPOINT ["/gateway", "serve"]
CMD ["/etc/cc-gateway/config.yaml"]
