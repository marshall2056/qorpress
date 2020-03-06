FROM golang:1.14-alpine AS builder
MAINTAINER x0rzkov

RUN apk add --no-cache make git openssl ca-certificates musl-dev gcc

COPY .  /go/src/github.com/qorpress/qorpress
WORKDIR /go/src/github.com/qorpress/qorpress

RUN cd /go/src/github.com/qorpress/qorpress \
 	&& go install

FROM alpine:3.11 AS runtime
MAINTAINER x0rzkov

ARG TINI_VERSION=${TINI_VERSION:-"v0.18.0"}

# Install tini to /usr/local/sbin
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini-muslc-amd64 /usr/local/sbin/tini

# Install runtime dependencies & create runtime user
RUN apk --no-cache --no-progress add ca-certificates \
	&& chmod +x /usr/local/sbin/tini && mkdir -p /opt \
	&& adduser -D qorpress -h /opt/qorpress -s /bin/sh \
	&& su qorpress -c 'cd /opt/qorpress; mkdir -p bin config'

# Switch to user context
# USER qorpress
WORKDIR /opt/qorpress/public

# copy executable
COPY --from=builder /go/bin/qorpress /opt/qorpress/bin/qorpress

ENV PATH $PATH:/opt/qorpress/bin

# Container configuration
EXPOSE 443 80 7000

VOLUME ["/opt/qorpress/public"]

ENTRYPOINT ["tini", "-g", "--"]
CMD ["/opt/qorpress/bin/qorpress"]
