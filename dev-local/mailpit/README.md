# Mailpit — local fake mail server

[Mailpit](https://mailpit.axllent.org/) is a fake SMTP server with a web UI.
Anything sent to its SMTP port shows up in a browser inbox, instead of actually being delivered.
This is useful for testing any CrownLabs component that sends real email (for example, tenant notifications) against a local cluster, without needing a real mail provider.

## 1. Deploy

We need to create:

- the `Deployment` of the mailpit container. The current setup:
    - does not require authentication, not considered required for development (can be changed with the `MP_SMTP_AUTH_ACCEPT_ANY` env variable);
    - does not store messages in a persistent way (can be changed by adding the `MP_DATABASE` env variable).
- the SMTP `Service` to send the emails. This Service can be contacted as `mailpit-smtp.default.svc.cluster.local:1025`.
- the HTTP `Service` to see the emails that have been sent.
- an `HTTPRoute` to have `mail.crownlabs.local:8443` resolve to the HTTP Service.

All these components can be set up with the following commands:

```bash
kubectl apply -f dev-local/mailpit/manifests
kubectl rollout status deployment/mailpit
```

## 2. Test it

The following command can be used to send a sample email:

```bash
kubectl run mailtest --rm -i --restart=Never --image=alpine --quiet -- sh -c '
apk add --no-cache netcat-openbsd >/dev/null 2>&1
printf "HELO test\r\nMAIL FROM:<test@crownlabs.local>\r\nRCPT TO:<john.doe@crownlabs.local>\r\nDATA\r\nSubject: Test\r\n\r\nHello from a test message.\r\n.\r\nQUIT\r\n" \
  | nc -w 3 mailpit-smtp.default.svc.cluster.local 1025
'
# Expected: "250 2.0.0 Ok: queued as <id>" in the output.
```

When you open [mail.crownlabs.local:8443](mail.crownlabs.local:8443), you should see that you correctly received the test e-mail.
No refresh of the website is needed, since new emails are pushed directly via WebSockets.

As an alternative to the web page, the emails can also be downloaded in JSON form, using the following command:

```bash
curl -sk https://mail.crownlabs.local:8443/api/v1/messages
```

## 3. Sending real emails (hint of)

The CrownLabs system does not have a high need of sending emails.
The main source of emails is the keycloak component, since a user needs to confirm its email address before being fully registered.
The Keycloak imported realm is already configured to correctly send emails, therefore you should already be able to use mailpit when creating a new user.

The other component that frequently needs to send emails is the `instance-automation`.
It is also already setup to work automatically, without any required setup.

If you need to develop a feature that needs to send emails, please refer to the `Crownmail` go library (inside the `operators` library)

## Next

- [`../envoy/README.md`](../envoy/README.md): the shared Gateway this routes through.
- [Mailpit's runtime options](https://mailpit.axllent.org/docs/configuration/runtime-options/), if you need to tweak retention, add basic authentication, and so on.
