# Tailscale

## Basic mode

Basic mode requires only a Tailscale auth key. Shadow Collective appears as a normal tailnet node named `shadow-collective` by default.

A configured gateway listener can then be reached by node name and port, for example:

```text
http://shadow-collective:10002
```

This mode is the easiest way to get started.

## Tailscale Services mode

Tailscale Services provide stable virtual service identities such as:

```text
svc:pihole-admin
svc:my-app
svc:postgres
```

A Service can be hosted by Shadow Collective while the actual backend stays on Fly's private network.

Tailscale currently requires a Service host to use a **tag-based identity**.

### 1. Define a gateway tag

Example policy fragment:

```json
{
  "tagOwners": {
    "tag:shadow-collective": ["autogroup:admin"]
  }
}
```

Create/tag the auth key appropriately, then set the Fly environment variable:

```toml
[env]
  TS_HOSTNAME = "shadow-collective"
  TS_TAGS = "tag:shadow-collective"
```

Redeploy after changing it.

### 2. Create Service resources in Tailscale

Create the desired Services in the Tailscale admin console. Service advertisement is an admin-controlled operation.

### 3. Create local gateway listeners

`config/services.json`:

```json
{
  "health_listen": "0.0.0.0:9000",
  "http": [
    {
      "name": "my-app",
      "listen": "0.0.0.0:10002",
      "upstream": "http://my-app.internal:8080"
    }
  ],
  "tcp": [],
  "dns": {
    "enabled": false,
    "listen": "0.0.0.0:53",
    "upstream": "shadow-pihole.internal:53"
  }
}
```

### 4. Map the Tailscale Service to the local listener

`config/tailscale-services.json`:

```json
{
  "version": "0.0.1",
  "services": {
    "svc:my-app": {
      "endpoints": {
        "tcp:443": "http://127.0.0.1:10002"
      }
    }
  }
}
```

At startup, Shadow Collective applies this file with `tailscale serve set-config --all` and attempts to advertise every Service in the file.

### 5. Deploy

```bash
fly deploy
```

If Tailscale requires approval, approve the Service host in the admin console or configure service auto-approval in your tailnet policy.

## Why Tailscale Services are optional

Services are great for stable identities and access policies, but they add a small amount of admin-side setup. The direct gateway-node/port mode continues to work without them.

## DNS is different

Tailscale Services currently map TCP endpoints. DNS commonly uses UDP, so Pi-hole DNS is not modeled as a normal Tailscale Service here.

Instead, tailnet DNS points directly to Shadow Collective's stable Tailscale node IP on port 53. The gateway forwards both UDP and TCP DNS to the private Pi-hole backend.

That is also why the Tailscale state volume matters.
