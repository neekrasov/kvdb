FROM golang:1.23.2-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /kvdb
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_HASH

RUN make build-server VERSION=$VERSION BUILD_TIME="$BUILD_TIME" GIT_HASH=$GIT_HASH

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata su-exec

COPY --from=builder /kvdb/bin/kvdb-server /usr/local/bin/kvdb-server
COPY config.yml /kvdb/etc/config.yml

COPY build/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

RUN addgroup -S kvdb && \
    adduser -S kvdb -G kvdb && \
    mkdir -p /var/lib/kvdb /var/log/kvdb && \
    chown -R kvdb:kvdb /var/lib/kvdb /var/log/kvdb /usr/local/bin/kvdb-server /usr/local/bin/docker-entrypoint.sh && \
    chmod 550 /usr/local/bin/kvdb-server

VOLUME ["/var/lib/kvdb", "/var/log/kvdb"]

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["kvdb-server", "run", "--config", "/kvdb/etc/config.yml"]

LABEL org.opencontainers.image.title="KVDB Server" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.created=$BUILD_TIME \
      org.opencontainers.image.revision=$GIT_HASH

