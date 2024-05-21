# Build e1s in go image
FROM golang:1.22.2-alpine as builder

# Set build argument for target architecture
ARG TARGETARCH

WORKDIR /src/e1s
COPY . ./

 # Set the GOARCH environment variable based on the TARGETARCH
 RUN GOARCH="" && \
     if [ "$TARGETARCH" = "amd64" ]; then \
       GOARCH="amd64"; \
     elif [ "$TARGETARCH" = "arm64" ]; then \
       GOARCH="arm64"; \
     else \
       echo "Unsupported architecture: $TARGETARCH"; \
       exit 1; \
     fi && \
     GOARCH=$GOARCH go mod vendor && \
     GOARCH=$GOARCH go build -o e1s ./cmd/e1s


# Install the session manager plugin in ubuntu image
# https://github.com/aws/session-manager-plugin/issues/12
FROM ubuntu:20.04 as sessionmanagerplugin

 # Install dependencies for dpkg and download tools
 RUN apt-get update && apt-get install -y curl dpkg && rm -rf /var/lib/apt/lists/*

 # Use the build argument to determine the correct package URL and installation method
 RUN if [ "$TARGETARCH" = "amd64" ]; then \
       curl -o session-manager-plugin.deb https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb; \
     elif [ "$TARGETARCH" = "arm64" ]; then \
       curl -o session-manager-plugin.deb https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_arm64/session-manager-plugin.deb; \
     else \
       echo "Unsupported architecture: $TARGETARCH"; \
       exit 1; \
     fi && \
     dpkg -i session-manager-plugin.deb && rm session-manager-plugin.deb

FROM alpine
COPY --from=builder /src/e1s/e1s /e1s
COPY --from=sessionmanagerplugin /usr/local/sessionmanagerplugin/bin/session-manager-plugin /usr/local/bin/


# Install the AWS CLI
RUN apk add --no-cache aws-cli \
		apk add --no-cache gcompat \
		rm -rf /var/cache/apk/*

ENTRYPOINT ["/e1s"]
