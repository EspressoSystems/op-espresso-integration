
FROM --platform=$BUILDPLATFORM golang:1.19.9-alpine3.16 as builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make git

COPY ./op-geth-proxy /app/op-geth-proxy
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app/op-geth-proxy

RUN go mod download

ARG TARGETOS TARGETARCH

RUN make op-geth-proxy VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.16

COPY --from=builder /app/op-geth-proxy/bin/op-geth-proxy /usr/local/bin

EXPOSE 9090

CMD ["op-geth-proxy"]
