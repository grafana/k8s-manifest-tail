# syntax=docker/dockerfile:1

FROM golang:1.25-alpine3.23 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/k8s-manifest-tail ./cmd

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /out/k8s-manifest-tail /usr/local/bin/k8s-manifest-tail
ENTRYPOINT ["/usr/local/bin/k8s-manifest-tail"]
