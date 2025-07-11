FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config/config.yaml ./config/config.yaml

RUN go build -o auth-service ./cmd

CMD ["./auth-service"]
