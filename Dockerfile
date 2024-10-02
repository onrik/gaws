FROM golang:1.23-alpine as builder

ADD . /gaws
WORKDIR /gaws

RUN go build -o /tmp/gaws .

FROM golang:1.23-alpine

RUN apk update
COPY --from=builder /tmp/gaws /usr/bin/gaws
