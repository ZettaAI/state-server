FROM golang:1.19-buster

WORKDIR /app
COPY . /app

RUN go build
CMD ./state-server
