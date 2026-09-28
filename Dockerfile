# Build stage
FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

# Production stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8000
CMD ["./app"]