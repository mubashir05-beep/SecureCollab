FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o auth ./cmd/auth

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/auth /app/auth
EXPOSE 8081
CMD ["/app/auth"]
