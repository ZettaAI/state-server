FROM golang:1.15.2-buster

WORKDIR /app
COPY . /app

RUN go get -v github.com/labstack/echo
