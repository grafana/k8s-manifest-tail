# syntax=docker/dockerfile:1@sha256:2780b5c3bab67f1f76c781860de469442999ed1a0d7992a5efdf2cffc0e3d769

FROM golang:1.26-alpine3.23@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS builder
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

FROM alpine:3.23@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40
WORKDIR /app
COPY --from=builder /out/k8s-manifest-tail /usr/local/bin/k8s-manifest-tail
RUN chmod +x /usr/local/bin/k8s-manifest-tail
ENTRYPOINT ["/usr/local/bin/k8s-manifest-tail"]
