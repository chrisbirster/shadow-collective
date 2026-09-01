# Security model

Shadow Collective is designed to keep private applications private by default.

## Trust boundaries

### Public internet

The gateway's `fly.toml` intentionally does not define a Fly public service. Do not add public port mappings unless you explicitly want public ingress.

### Tailnet

Tailscale authenticates devices and users. Tailnet grants/ACLs should decide which users/devices can reach Shadow Collective and individual Tailscale Services.

### Fly organization network

Once traffic reaches Shadow Collective, approved routes can connect to backend applications on the Fly organization's private 6PN. Treat membership in that Fly private network as another trust boundary.

## Secrets

Never commit:

- `TAILSCALE_AUTHKEY`
- Pi-hole admin passwords
- Fly API tokens
- private keys
- OAuth client secrets

Use Fly secrets:

```bash
fly secrets set NAME=value
```

The repository contains only secret **names** and examples.

## Tailscale state

`/data/tailscaled.state` is sensitive because it contains the node's Tailscale identity. It is kept on a Fly Volume and is not copied into the repository or Docker image.

## DNS

The gateway starts Tailscale with `--accept-dns=false`. This prevents the gateway itself from consuming tailnet DNS settings that may point back to the gateway.

Only enable DNS forwarding after the configured upstream resolver exists and is reachable.

## Tailscale Services

Service hosts require tag-based identity. Use narrow tags and grants rather than granting broad access to the entire gateway node when practical.

A useful end state is:

```text
user/device -> specific svc:resource -> Shadow Collective -> specific private backend
```

rather than unrestricted access to every port on the gateway.

## Public repository safety

This repository is safe to keep public as long as configuration contains only non-secret infrastructure names and secrets remain in Fly/Tailscale secret stores.

Hostnames and architecture can still reveal topology. If a backend name itself is sensitive, use generic aliases in the public config and keep deployment-specific configuration on a private branch or external secret/config system.
