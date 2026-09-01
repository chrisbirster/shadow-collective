# Networking

Shadow Collective bridges two private networks without exposing application services to the public internet.

## Network A: Tailscale

This is the client-facing private network. Macs, phones, Apple TVs, laptops, and other authorized devices join the tailnet.

The gateway receives a Tailscale IP and MagicDNS name such as:

```text
shadow-collective
shadow-collective.<tailnet>.ts.net
```

The exact tailnet suffix is assigned by Tailscale.

## Network B: Fly 6PN

Fly apps inside the same Fly organization are connected using Fly's private WireGuard-based IPv6 network. Service discovery is available through `.internal` names.

Examples:

```text
shadow-pihole.internal
my-api.internal
postgres.internal
```

The Go gateway lets the operating system resolve these names and dial the resulting private Fly address.

## No public ingress by default

The root `fly.toml` has no Fly `services` or `http_service` declaration. That is intentional.

Do not add a public Fly service just to make a backend reachable through Shadow Collective. The path should remain:

```text
Tailscale -> gateway -> Fly 6PN -> backend
```

You can audit the gateway app with:

```bash
fly ips list
fly config show
```

## Ports

The gateway configuration allocates local listener ports. Example convention:

| Range | Suggested use |
|---|---|
| `53` | DNS forwarding |
| `9000` | gateway health |
| `10000-10999` | HTTP service proxies |
| `15000-15999` | raw TCP services |

These ranges are conventions, not protocol requirements.

## DNS

When DNS forwarding is enabled, both UDP and TCP listeners are started on the configured address. Supporting both matters because normal DNS often uses UDP but can fall back to TCP.

Recommended config:

```json
"dns": {
  "enabled": true,
  "listen": "0.0.0.0:53",
  "upstream": "shadow-pihole.internal:53"
}
```

Then configure the gateway's Tailscale IP as a nameserver in the Tailscale admin DNS settings.

## Apple TV

Apple TV can run the Tailscale tvOS client. Once it participates in the tailnet and accepts the tailnet DNS settings, its DNS requests can traverse Shadow Collective to Pi-hole just like requests from your computers and phones.

Pi-hole remains DNS-level filtering; it cannot reliably identify YouTube video ads that share delivery infrastructure with normal YouTube video segments.
