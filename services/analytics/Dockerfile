FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o analytics ./cmd/analytics

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/analytics /app/analytics
EXPOSE 8084
CMD ["/app/analytics"]
