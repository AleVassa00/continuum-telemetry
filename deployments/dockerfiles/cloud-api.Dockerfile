FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/cloud-api ./cmd/cloud-api

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /out/cloud-api /app/cloud-api
COPY --from=builder /src/config /app/config

EXPOSE 8080

ENTRYPOINT ["/app/cloud-api"]