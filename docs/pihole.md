# Pi-hole on Fly.io

Pi-hole is an optional companion application. It does **not** run Tailscale. Shadow Collective reaches it over Fly's private network.

An example configuration lives at:

```text
deploy/pihole/fly.toml
```

## 1. Create the Pi-hole app

The example uses `shadow-pihole`. Fly app names are globally unique; change it if required.

```bash
fly apps create shadow-pihole
```

Keep it in the **same Fly organization/private network** as Shadow Collective so `shadow-pihole.internal` resolves and is reachable.

## 2. Set the Pi-hole web password

Pi-hole's Docker configuration supports its admin password through `FTLCONF_webserver_api_password`.

Store it as a Fly secret:

```bash
fly secrets set \
  -a shadow-pihole \
  FTLCONF_webserver_api_password='use-a-long-random-password'
```

## 3. Deploy Pi-hole

From the repository root:

```bash
fly deploy -c deploy/pihole/fly.toml
```

The example declares a persistent volume at `/etc/pihole` and enables Pi-hole's DNS listening mode for private network access.

Do not add a public Fly `http_service` or DNS service to the Pi-hole configuration.

## 4. Enable DNS forwarding in Shadow Collective

Update `config/services.json`:

```json
"dns": {
  "enabled": true,
  "listen": "0.0.0.0:53",
  "upstream": "shadow-pihole.internal:53"
}
```

Then:

```bash
fly deploy
```

## 5. Point Tailscale DNS to Shadow Collective

Find the gateway's Tailscale IP:

```bash
fly ssh console -C '/app/tailscale --socket=/var/run/tailscale/tailscaled.sock ip -4'
```

In the Tailscale admin DNS settings, add that address as the nameserver you want tailnet clients to use. Enable DNS override if your desired policy requires every tailnet device to use it.

Flow:

```text
client DNS query
   -> gateway Tailscale IP:53
   -> shadow-pihole.internal:53
   -> Pi-hole blocklists/cache
   -> upstream DNS
```

## Optional: expose the Pi-hole admin UI privately

Add an HTTP route:

```json
{
  "name": "pihole-admin",
  "listen": "0.0.0.0:10001",
  "upstream": "http://shadow-pihole.internal:80"
}
```

Then access it through Shadow Collective from a tailnet device or attach a Tailscale Service such as `svc:pihole-admin`.

## What Pi-hole will and will not block

Pi-hole works at DNS level. It is effective for many advertising networks, telemetry hosts, trackers, malicious domains, and unwanted third-party services.

It cannot reliably remove YouTube pre-roll and mid-roll video ads from the official YouTube app because DNS does not tell Pi-hole whether a particular HTTPS request to shared YouTube infrastructure is an advertisement or requested video content.
