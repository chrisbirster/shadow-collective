# Deployment

## Requirements

- A Fly.io account with `flyctl` installed and authenticated.
- A Tailscale tailnet.
- A Tailscale auth key stored as a Fly secret.

## One-time bootstrap

### Create the Fly app

The repository ships with:

```toml
app = "shadow-collective"
```

Fly app names are globally unique. If that name is unavailable, change it in `fly.toml` before creating the app.

```bash
fly apps create shadow-collective
```

### Set the Tailscale secret

```bash
fly secrets set TAILSCALE_AUTHKEY='tskey-...'
```

Do not put the key in `fly.toml`, Docker build arguments, `.env` files committed to Git, or gateway JSON.

### First deploy

```bash
fly deploy
```

`fly.toml` declares a 1 GB `shadow_state` volume with `initial_size`, so Fly can provision the persistent state volume on the first deployment.

## Normal deployments

Once bootstrapped:

```bash
fly deploy
```

That is the intended operational loop.

## Verify

```bash
fly status
fly logs
```

Check Tailscale from inside the Machine:

```bash
fly ssh console -C '/app/tailscale --socket=/var/run/tailscale/tailscaled.sock status'
```

Check the gateway process:

```bash
fly ssh console -C 'wget -qO- http://127.0.0.1:9000/healthz'
```

Expected:

```text
ok
```

## Updating routes

Edit:

```text
config/services.json
```

Then:

```bash
fly deploy
```

For Tailscale Service identities, also edit:

```text
config/tailscale-services.json
```

and redeploy.

## Region

The default region is `iad`. Keeping the gateway and its backend services in the same or nearby Fly region reduces latency, especially for DNS.

Change `primary_region` if your workloads live elsewhere.

## Resource size

The default is:

```toml
[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"
```

That should be enough for a small gateway. Increase memory if you configure many concurrent proxy connections or observe pressure.

## Rollback

Because application configuration is in Git, revert the bad change and redeploy:

```bash
git revert <commit>
fly deploy
```

The Tailscale identity remains on the persistent volume.
