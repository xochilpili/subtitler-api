FROM golang:1.24-alpine AS builder
RUN apk add build-base git openssh-client openssl-dev librdkafka-dev librdkafka pkgconf

RUN mkdir /app

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN go build -tags musl -o /app/subtitler-api cmd/main.go

FROM alpine:3.19
LABEL MAINTAINER="xochilpili <xochilpili@gmail.com>"
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/subtitler-api ./subtitler-api

CMD ["./subtitler-api"]