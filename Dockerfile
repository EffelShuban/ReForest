FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/auth-service ./cmd/auth
RUN go build -o /bin/api-gateway ./cmd/api

FROM alpine:latest AS auth-service
WORKDIR /root/
COPY --from=builder /bin/auth-service .
EXPOSE 50051
CMD ["./auth-service"]

FROM alpine:latest AS api-gateway
WORKDIR /root/
COPY --from=builder /bin/api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]