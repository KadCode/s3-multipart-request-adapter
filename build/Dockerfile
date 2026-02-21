FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o s3-adapter main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/s3-adapter .

COPY certs ./certs
COPY config/config.yaml ./config/config.yaml

EXPOSE 8080

CMD ["./s3-adapter"]
