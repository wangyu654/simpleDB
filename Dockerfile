FROM golang:latest AS builder

# MAINTAINER wangyu

ENV GOPROXY https://goproxy.cn
COPY go.mod .
COPY go.sum .
RUN go mod download

WORKDIR /simpleDB

COPY . ./

RUN go build .

EXPOSE 8000

ENTRYPOINT  ["./simpleDB"]
