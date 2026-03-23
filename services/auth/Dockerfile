FROM golang:1.22-alpine AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o auth ./cmd/auth

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/auth /app/auth
EXPOSE 8081
CMD ["/app/auth"]
