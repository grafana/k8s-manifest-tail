# syntax=docker/dockerfile:1@sha256:87999aa3d42bdc6bea60565083ee17e86d1f3339802f543c0d03998580f9cb89

FROM golang:1.27-alpine3.23@sha256:cee7b6811ec0b093ff42b3a97ddc48d1ab0258e35e7a80643fd4e32a8a610cb3 AS builder
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

FROM alpine:3.24@sha256:ce9e1e6d23f04a0b7593445a062a9fc73ed52f8260b9396f77a958295ad18ee7
WORKDIR /app
COPY --from=builder /out/k8s-manifest-tail /usr/local/bin/k8s-manifest-tail
RUN chmod +x /usr/local/bin/k8s-manifest-tail
ENTRYPOINT ["/usr/local/bin/k8s-manifest-tail"]
