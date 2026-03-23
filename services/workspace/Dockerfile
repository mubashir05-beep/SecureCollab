FROM golang:1.22-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /workspace ./cmd/workspace

FROM alpine:3.19
COPY --from=builder /workspace /workspace
EXPOSE 8086
CMD ["/workspace"]
