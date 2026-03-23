FROM golang:1.22-alpine AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gateway ./cmd/gateway

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/gateway /app/gateway
EXPOSE 8080
CMD ["/app/gateway"]
