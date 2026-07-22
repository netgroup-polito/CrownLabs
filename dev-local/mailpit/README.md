# Mailpit — local fake mail server

[Mailpit](https://mailpit.axllent.org/) is a fake SMTP server + web UI: anything sent to
its SMTP port shows up in a browser inbox instead of actually being delivered. Useful
for testing any CrownLabs component that sends real email (e.g. tenant notifications)
against a local cluster, without needing a real mail provider.

## Prerequisites

- The base k3s cluster — see [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md) set up (Gateway API CRDs, the controller, the
  shared `crownlabs` Gateway) — Mailpit's web UI is exposed entirely through the
  `HTTPRoute` in `manifests/mailpit.yaml`, which needs that `Gateway` to already exist.

## Deploy

```bash
kubectl apply -f dev-local/mailpit/manifests
kubectl rollout status deployment/mailpit
```

This creates:
- The `mailpit` `Deployment` (image `axllent/mailpit`). No persistence — Mailpit
  defaults to an in-memory message store when `MP_DATABASE` isn't set, same ephemeral
  tradeoff already made for Postgres in `../keycloak/manifests/keycloak.yaml`. No auth
  (`MP_SMTP_AUTH_ACCEPT_ANY=1`) — this is local dev only.
- `Service mailpit-smtp` (port 1025) — **ClusterIP only, not exposed through the
  Gateway**. SMTP isn't HTTP, so it can't go behind an `HTTPRoute`; this is consumed
  in-cluster by whatever CrownLabs component ends up sending mail
  (`mailpit-smtp.default.svc.cluster.local:1025`).
- `Service mailpit` (port 8025, the web UI) + an `HTTPRoute` routing
  `mail.crownlabs.local` to it through the shared `crownlabs` Gateway — same pattern as
  [`../keycloak/manifests/keycloak.yaml`](../keycloak/manifests/keycloak.yaml)'s own
  `HTTPRoute`.

## `/etc/hosts`

Add the hostname alongside Keycloak's (see [`../envoy/README.md`](../envoy/README.md)
step 4 — same WSL2 caveat applies: the Windows-side hosts file needs this too, not just
WSL2's):

```bash
echo "127.0.0.1 mail.crownlabs.local" | sudo tee -a /etc/hosts
```

(WSL2 browser: `Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 mail.crownlabs.local"` from an elevated PowerShell.)

## Test it

**Send a test message** — from anywhere inside the cluster (no auth, no TLS on the SMTP
side):

```bash
kubectl run mailtest --rm -i --restart=Never --image=alpine --quiet -- sh -c '
apk add --no-cache netcat-openbsd >/dev/null 2>&1
printf "HELO test\r\nMAIL FROM:<test@crownlabs.local>\r\nRCPT TO:<john.doe@crownlabs.local>\r\nDATA\r\nSubject: Test\r\n\r\nHello from a test message.\r\n.\r\nQUIT\r\n" \
  | nc -w 3 mailpit-smtp.default.svc.cluster.local 1025
'
# expected: "250 2.0.0 Ok: queued as <id>" in the output
```

**See it arrive** — open `https://mail.crownlabs.local` (WSL2: append `:8443`, same
bridge as Keycloak, see [`../envoy/README.md`](../envoy/README.md) step 5) in a browser:
the message shows up in the inbox immediately, no refresh needed (Mailpit pushes new
messages over websocket/SSE). Or hit the API directly:

```bash
curl -sk https://mail.crownlabs.local/api/v1/messages   # WSL2: :8443
```

## Next

- [`../envoy/README.md`](../envoy/README.md) — the shared Gateway this routes through.
- [Mailpit's runtime options](https://mailpit.axllent.org/docs/configuration/runtime-options/)
  if you need to tweak retention, add basic auth, etc.
