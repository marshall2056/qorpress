FROM golang:1.14-alpine AS builder

RUN apk add --no-cache make git openssl ca-certificates

COPY .  /go/src/github.com/qorpress/qorpress
WORKDIR /go/src/github.com/qorpress/qorpress

RUN cd /go/src/github.com/qorpress/qorpress \
	&& go get -v \
 	&& go install

EXPOSE 7000

CMD ["/go/bin/qorpress-example"]