# Multi-stage build — final image < 15MB
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /gateway ./cmd/gateway

FROM scratch
COPY --from=builder /gateway /gateway
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8443
ENTRYPOINT ["/gateway", "serve"]
CMD ["/etc/cc-gateway/config.yaml"]
