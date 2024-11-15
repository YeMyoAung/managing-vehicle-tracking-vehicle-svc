# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY main.go .
COPY go.mod .
COPY go.sum .
COPY internal/ internal/

RUN go mod tidy && go build -o bin/vehicle-svc

# Stage 2: Set up the final image
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/bin/vehicle-svc .
COPY .env.uat .env

EXPOSE 10001

CMD ["./vehicle-svc"]
