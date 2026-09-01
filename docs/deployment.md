# Deployment

## Requirements

- A Fly.io account with `flyctl` installed and authenticated.
- A Tailscale tailnet.
- A non-ephemeral Tailscale auth key for first enrollment.

The auth key is a **bootstrap credential only**. After enrollment, Shadow Collective persists its Tailscale node identity on the Fly Volume at `/data/tailscaled.state`; normal restarts and deploys do not require the auth key.

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

### Set the Tailscale enrollment secret

Avoid typing the auth key directly into a shell command because it can end up in shell history. On zsh, a safe interactive pattern is:

```zsh
read -rs "TAILSCALE_AUTHKEY?Tailscale auth key: "
echo
fly secrets set -a shadow-collective TAILSCALE_AUTHKEY="$TAILSCALE_AUTHKEY"
unset TAILSCALE_AUTHKEY
```

Do not put the key in `fly.toml`, Docker build arguments, committed `.env` files, or gateway JSON.

### First deploy

```bash
fly deploy
```

`fly.toml` declares a 1 GB `shadow_state` volume. The Tailscale identity is stored there so it survives Machine replacements.

### Remove the enrollment key after verification

Once the node appears online in Tailscale and `/data/tailscaled.state` exists, the enrollment key is no longer required for normal operation. Revoke the key in the Tailscale admin console and remove it from Fly:

```bash
fly secrets unset -a shadow-collective TAILSCALE_AUTHKEY
```

Revoking an auth key does not deauthorize a node that was already enrolled with it.

## Normal deployments

Once bootstrapped:

```bash
fly deploy
```

That is the intended operational loop. The persistent Tailscale state is the source of truth for node identity.

## Verify

```bash
fly status
fly logs
```

Check Tailscale from inside the Machine:

```bash
fly ssh console -C '/app/tailscale --socket=/var/run/tailscale/tailscaled.sock status'
```

Check that durable state exists:

```bash
fly ssh console -C 'ls -lh /data/tailscaled.state'
```

Check the gateway process:

```bash
fly ssh console -C 'wget -qO- http://127.0.0.1:9000/healthz'
```

Expected:

```text
ok
```

## Re-enrollment

If the node is explicitly logged out or `/data/tailscaled.state` is removed, create a fresh auth key and set `TAILSCALE_AUTHKEY` again before restarting or deploying. The startup script requires an auth key only when no persisted node state is available.

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
