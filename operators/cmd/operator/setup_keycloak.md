- client -> settings -> service account roles (enabled), Standard flow (disabled)
- client -> Service account roles -> Assign role -> (realm-management) manage-users, manage-events, view-events, query-clients, view-clients, manage-clients
  (if you don't see the user, go to Clients -> crownlabs -> service account roles -> link to service-account-crownlabs)

**NOTE**: the client MUST be different from the one used to sign in the users (frontend and k8s API gateway).


# Webhooks
- Realm Settings -> Events -> add ext-event-webhook

- send a POST request to: http:// keycloak-endpoint/realms/crownlabs/webhooks

```json
{
  "enabled": "true",
  "url": "http:// operator-endpoint:8082/tenant-webhook",
  "eventTypes": [
    "admin.USER-UPDATE",
    "access.CUSTOM_REQUIRED_ACTION"
  ]
}
```
