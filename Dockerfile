FROM golang:1.24-alpine AS builder

ADD . /gaws
WORKDIR /gaws

RUN go build -o /tmp/gaws .

FROM golang:1.24-alpine

RUN apk update
COPY --from=builder /tmp/gaws /usr/bin/gaws
