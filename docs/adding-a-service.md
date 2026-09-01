# Adding a service

The intended workflow is config-first.

## HTTP service

Assume a Fly app named `notes` listens privately on port `8080`.

### Backend requirement

The backend should be reachable over Fly's private network as:

```text
notes.internal:8080
```

It does not need Tailscale.

### Add a gateway route

Edit `config/services.json`:

```json
{
  "name": "notes",
  "listen": "0.0.0.0:10003",
  "upstream": "http://notes.internal:8080"
}
```

Add that object to the `http` array.

Deploy:

```bash
fly deploy
```

Basic access:

```text
http://shadow-collective:10003
```

## Stable Tailscale Service name

If you have enabled Tailscale Services, add a mapping to `config/tailscale-services.json`:

```json
"svc:notes": {
  "endpoints": {
    "tcp:443": "http://127.0.0.1:10003"
  }
}
```

Create the corresponding Service resource in Tailscale and redeploy.

## TCP service

For PostgreSQL:

```json
{
  "name": "postgres",
  "listen": "0.0.0.0:15432",
  "upstream": "postgres.internal:5432"
}
```

Add it to the `tcp` array and redeploy.

The gateway does not inspect the TCP payload; it simply creates a bidirectional stream between the tailnet-side listener and the private Fly backend.

## Checklist

1. Backend app is in the same Fly organization/private network.
2. Backend listens on its Fly private interface / appropriate IPv6 wildcard.
3. No public Fly service is required solely for Shadow Collective.
4. Choose an unused gateway listener port.
5. Add the upstream to `config/services.json`.
6. Optionally create a Tailscale Service mapping.
7. `fly deploy`.
8. Test from a tailnet client.
