FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/golang-test-task

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/app .

COPY --from=builder /app/config.yaml ./config.default.yaml

EXPOSE 8080

# Run the application
CMD ["./app", "-config", "/app/config/config.yaml"]
