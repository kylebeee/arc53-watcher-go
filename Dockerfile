# syntax = docker/dockerfile:experimental

# BUILD
FROM golang:1.21-alpine as builder

RUN apk add --no-cache git

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app main/*.go

# PACKAGE
FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

COPY --from=builder ./app .

EXPOSE 5000/tcp

CMD ["./app"]