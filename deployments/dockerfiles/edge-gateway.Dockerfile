FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/edge-gateway ./cmd/edge-gateway

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/edge-gateway /app/edge-gateway
COPY --from=builder /src/config /app/config

EXPOSE 8081

ENTRYPOINT ["/app/edge-gateway"]