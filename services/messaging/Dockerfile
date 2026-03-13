FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o messaging ./cmd/messaging

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/messaging /app/messaging
EXPOSE 8083
CMD ["/app/messaging"]
