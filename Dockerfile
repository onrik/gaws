FROM golang:1.26-alpine AS builder

ADD . /gaws
WORKDIR /gaws

RUN go build -o /tmp/gaws .

FROM golang:1.26-alpine

RUN apk update
COPY --from=builder /tmp/gaws /usr/bin/gaws
