FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY config/config.yaml ./config/config.yaml

RUN go build -o auth-service ./cmd

CMD ["./auth-service"]
