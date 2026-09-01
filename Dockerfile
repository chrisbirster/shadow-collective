# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/shadow-collective ./cmd/shadow-collective

FROM alpine:3.22
RUN apk add --no-cache ca-certificates iptables ip6tables jq

COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscaled /app/tailscaled
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscale /app/tailscale
COPY --from=build /out/shadow-collective /app/shadow-collective
COPY config /app/config
COPY scripts/start.sh /app/start.sh

RUN mkdir -p /data /var/run/tailscale && chmod +x /app/start.sh

ENTRYPOINT ["/app/start.sh"]
