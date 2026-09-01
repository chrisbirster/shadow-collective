# Architecture

## Goal

Shadow Collective is a small private-cloud edge for a Fly.io organization. Instead of embedding Tailscale into every application, one dedicated Fly app owns the Tailscale identity and proxies approved traffic to private backends.

## Components

### Shadow Collective gateway

Runs on one Fly Machine and contains:

1. `tailscaled` — joins the Machine to the tailnet.
2. `shadow-collective` — a small Go proxy that reads `config/services.json`.
3. Optional Tailscale Services configuration — `config/tailscale-services.json`.
4. A persistent Fly Volume at `/data` containing Tailscale state.

### Fly private network

Fly apps in the same organization can communicate over Fly's private IPv6 6PN network. Fly provides `.internal` DNS names such as:

```text
my-app.internal
shadow-pihole.internal
postgres.internal
```

Shadow Collective uses those names as upstreams. Backend applications therefore need no Tailscale SDK, daemon, auth key, or public ingress.

### Tailnet clients

Devices such as macOS, iPhone, iPad, Apple TV, Linux, and Windows connect to Tailscale normally. They see Shadow Collective as another tailnet node and, optionally, see stable Tailscale Services advertised by it.

## Traffic flows

### HTTP

```text
Tailnet device
    -> Shadow Collective Tailscale interface
    -> local Go HTTP reverse proxy
    -> app.internal:port over Fly 6PN
```

### TCP

```text
Tailnet device
    -> Shadow Collective
    -> local TCP proxy
    -> database.internal:5432 over Fly 6PN
```

### DNS / Pi-hole

```text
Tailnet device
    -> Shadow Collective:53 (UDP or TCP)
    -> shadow-pihole.internal:53
    -> Pi-hole
    -> upstream resolver
```

`tailscaled` starts with `--accept-dns=false`. This is deliberate: if the tailnet's global DNS server is Shadow Collective itself, allowing the gateway to consume that same DNS configuration could create a resolver loop.

## Why a persistent volume?

DNS clients need a stable Tailscale target. `/data/tailscaled.state` stores the gateway node identity. A Fly Machine can be replaced during a deploy while the persistent state remains attached, so the node does not need to become a brand-new Tailscale identity on every release.

The same volume also avoids continuously consuming an auth key during normal restarts.

## Why Pi-hole is separate

Pi-hole is intentionally a separate Fly app:

- Pi-hole can be upgraded or replaced independently.
- Shadow Collective stays a general-purpose gateway.
- The Pi-hole admin UI can itself be exposed through Shadow Collective.
- Pi-hole has its own persistent state and lifecycle.
- Other private Fly services can use the same gateway pattern.

## Why there is no Fly public service

The root `fly.toml` intentionally contains no `[[services]]` or `[http_service]` section. Traffic reaches the gateway through Tailscale rather than Fly Proxy. Backend services should follow the same private-by-default model unless they independently need public ingress.

## Tailscale Services

Tailscale Services add a virtual service identity such as `svc:my-app` in front of a local proxy listener on Shadow Collective.

```text
svc:my-app
    -> Shadow Collective
    -> 127.0.0.1:10002
    -> my-app.internal:8080
```

This decouples the client-visible identity from the physical gateway node. See `tailscale.md` for setup and approval requirements.

## Scaling

Start with one gateway Machine. DNS UDP forwarding currently targets the gateway node directly, and the persisted node identity is simple and predictable.

For higher availability later:

- HTTP/TCP services can use multiple Tailscale Service hosts.
- DNS can move behind a dedicated highly available resolver strategy.
- Pi-hole can be replaced by redundant DNS resolvers if availability becomes more important than simplicity.
