FROM golang:1.22-alpine AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o keydist ./cmd/keydist

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/keydist /app/keydist
EXPOSE 8082
CMD ["/app/keydist"]
