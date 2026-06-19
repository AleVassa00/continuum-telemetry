FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/sensor-simulator ./cmd/sensor-simulator

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/sensor-simulator /app/sensor-simulator
COPY --from=builder /src/config /app/config

ENTRYPOINT ["/app/sensor-simulator"]