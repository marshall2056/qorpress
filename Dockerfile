FROM golang:1.14-alpine AS builder

RUN apk add --no-cache make git openssl ca-certificates musl-dev gcc

COPY .  /go/src/github.com/qorpress/qorpress
WORKDIR /go/src/github.com/qorpress/qorpress

RUN cd /go/src/github.com/qorpress/qorpress \
 	&& go install

EXPOSE 7000 443 80

CMD ["/go/bin/qorpress"]