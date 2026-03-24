# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.1
ARG ALPINE_VERSION=3.23

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /src

RUN apk add --no-cache upx ca-certificates tzdata

# Cache deps first.
COPY go.mod go.sum ./
RUN go mod download

# Build sources.
COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN go tool task build:bins:upx GOOS=${TARGETOS} GOARCH=${TARGETARCH} && \
  if [ "${TARGETOS}" = "windows" ]; then \
    cp /src/dist/server.exe /out/server && \
    cp /src/dist/migrate.exe /out/migrate ; \
  else \
    cp /src/dist/server /out/server && \
    cp /src/dist/migrate /out/migrate ; \
  fi

FROM scratch AS server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]

FROM scratch AS migrate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/migrate /migrate

ENTRYPOINT ["/migrate"]
