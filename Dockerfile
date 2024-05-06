# Build e1s in go image
FROM golang:1.22.2-alpine as builder

WORKDIR /src/e1s
COPY . ./
RUN go mod vendor && \
    go build -o e1s ./cmd/e1s


# Install the session manager plugin in ubuntu image
# https://github.com/aws/session-manager-plugin/issues/12
FROM ubuntu:20.04 as sessionmanagerplugin

ADD https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb .
RUN dpkg -i "session-manager-plugin.deb"

FROM alpine
COPY --from=builder /src/e1s/e1s /e1s
COPY --from=sessionmanagerplugin /usr/local/sessionmanagerplugin/bin/session-manager-plugin /usr/local/bin/


# Install the AWS CLI
RUN apk add --no-cache aws-cli \
		apk add --no-cache gcompat \
		rm -rf /var/cache/apk/*

ENTRYPOINT ["/e1s"]