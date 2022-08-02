FROM golang:1.18-alpine as builder

ADD . /gaws
WORKDIR /gaws

RUN go build -o /tmp/gaws .

FROM golang:1.18-alpine

RUN apk update
COPY --from=builder /tmp/gaws /usr/bin/gaws
