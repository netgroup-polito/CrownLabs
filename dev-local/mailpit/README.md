# Mailpit — local fake mail server

[Mailpit](https://mailpit.axllent.org/) is a fake SMTP server with a web UI.
Anything sent to its SMTP port shows up in a browser inbox, instead of actually being delivered.
This is useful for testing any CrownLabs component that sends real email (for example, tenant notifications) against a local cluster, without needing a real mail provider.

## Prerequisites

- The base k3s cluster. See [`../base-k3s/README.md`](../base-k3s/README.md).
- [Envoy Gateway](../envoy/README.md) set up: the Gateway API CRDs, the controller, and the
  shared `crownlabs` Gateway.
  - Mailpit's web UI exposed entirely through the
  `HTTPRoute` in `manifests/mailpit.yaml`, which needs that `Gateway` to already exist.

## Deploy

```bash
kubectl apply -f dev-local/mailpit/manifests
kubectl rollout status deployment/mailpit
```

This creates:
- The `mailpit` `Deployment` (image `axllent/mailpit`). It has no persistence: Mailpit defaults to an in-memory message store when `MP_DATABASE` is not set, the same ephemeral tradeoff already made for Postgres in `../keycloak/manifests/keycloak.yaml`. It has no authentication (`MP_SMTP_AUTH_ACCEPT_ANY=1`), since this is local development only.
- `Service mailpit-smtp` (port 1025). This is **ClusterIP only, and not exposed through the Gateway**. SMTP is not HTTP, so it cannot go behind an `HTTPRoute`. Whatever CrownLabs component ends up sending mail consumes this in-cluster, at `mailpit-smtp.default.svc.cluster.local:1025`.
- `Service mailpit` (port 8025, the web UI), and an `HTTPRoute` that routes `mail.crownlabs.local` to it through the shared `crownlabs` Gateway. This is the same pattern used by [`../keycloak/manifests/keycloak.yaml`](../keycloak/manifests/keycloak.yaml)'s own `HTTPRoute`.

## `/etc/hosts`

Add the hostname alongside Keycloak's.
See [`../envoy/README.md`](../envoy/README.md) step 4: the same WSL2 caveat applies, and the Windows-side hosts file needs this too, not just WSL2's.

```bash
echo "127.0.0.1 mail.crownlabs.local" | sudo tee -a /etc/hosts
```

On the WSL2 browser, run this from an elevated PowerShell instead:
```powershell
Add-Content -Path C:\Windows\System32\drivers\etc\hosts -Value "127.0.0.1 mail.crownlabs.local"
```

## Test it

**Send a test message**, from anywhere inside the cluster. There is no authentication, and no TLS, on the SMTP side:

```bash
kubectl run mailtest --rm -i --restart=Never --image=alpine --quiet -- sh -c '
apk add --no-cache netcat-openbsd >/dev/null 2>&1
printf "HELO test\r\nMAIL FROM:<test@crownlabs.local>\r\nRCPT TO:<john.doe@crownlabs.local>\r\nDATA\r\nSubject: Test\r\n\r\nHello from a test message.\r\n.\r\nQUIT\r\n" \
  | nc -w 3 mailpit-smtp.default.svc.cluster.local 1025
'
# Expected: "250 2.0.0 Ok: queued as <id>" in the output.
```

**See it arrive.** Open `https://mail.crownlabs.local` in a browser (on WSL2, append `:8443`; this is the same bridge used for Keycloak, see [`../envoy/README.md`](../envoy/README.md) step 5).
The message shows up in the inbox immediately, with no refresh needed, since Mailpit pushes new messages over WebSocket/SSE.
Or hit the API directly:

```bash
curl -sk https://mail.crownlabs.local/api/v1/messages   # On WSL2: :8443
```

## Next

- [`../envoy/README.md`](../envoy/README.md): the shared Gateway this routes through.
- [Mailpit's runtime options](https://mailpit.axllent.org/docs/configuration/runtime-options/), if you need to tweak retention, add basic authentication, and so on.
