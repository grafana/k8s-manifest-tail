# syntax=docker/dockerfile:1@sha256:87999aa3d42bdc6bea60565083ee17e86d1f3339802f543c0d03998580f9cb89

FROM golang:1.27-alpine3.23@sha256:9e45f0eb4a63ed37ad6e604950407558b7c738e34e015ae77a5d4ad369cde079 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION=dev
ARG GIT_COMMIT=""
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags "-X github.com/grafana/k8s-manifest-tail/cmd.version=${VERSION} -X github.com/grafana/k8s-manifest-tail/cmd.gitCommit=${GIT_COMMIT}" \
    -o /out/k8s-manifest-tail .

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
WORKDIR /app
COPY --from=builder /out/k8s-manifest-tail /usr/local/bin/k8s-manifest-tail
RUN chmod +x /usr/local/bin/k8s-manifest-tail
ENTRYPOINT ["/usr/local/bin/k8s-manifest-tail"]
