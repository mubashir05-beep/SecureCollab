FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o analytics ./cmd/analytics

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/analytics /app/analytics
EXPOSE 8084
CMD ["/app/analytics"]
