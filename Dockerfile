# golang alpine 1.13.x
FROM golang:1.13-alpine as builder

LABEL maintainer="wout.slakhorst@nuts.nl"

RUN apk update \
 && apk add --no-cache \
            gcc=9.2.0-r4 \
            musl-dev=1.1.24-r2 \
 && update-ca-certificates

ENV GO111MODULE on
ENV GOPATH /

RUN mkdir /opt/nuts && cd /opt/nuts
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /opt/nuts/nuts-network

# alpine 3.11.x
FROM alpine:3.11
RUN apk update \
  && apk add --no-cache \
             ca-certificates=20191127-r1 \
             tzdata \
             curl \
  && update-ca-certificates
COPY --from=builder /opt/nuts/nuts-network /usr/bin/nuts-network

HEALTHCHECK --start-period=30s --timeout=5s --interval=10s \
    CMD curl -f http://localhost:1323/status || exit 1

EXPOSE 1323 5555
CMD ["/bin/sh", "-c", "/usr/bin/nuts-network server --publicAddr=$HOSTNAME:5555 --nodeId=$HOSTNAME"]
