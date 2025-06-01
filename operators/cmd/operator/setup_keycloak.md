- client -> settings -> service account roles (enabled)
- users -> service-account-crownlabs -> Role mapping -> Assign role -> (realm-management) manage-users
  (if you don't see the user, go to Clients -> crownlabs -> service account roles -> link to service-account-crownlabs)


# Webhooks
- Realm Settings -> Events -> add ext-event-webhook

- send a POST request to: http://keycloak-endpoint/realms/crownlabs/webhooks

```json
{
  "enabled": "true",
  "url": "http://operator-endpoint:8082/tenant-webhook",
  "eventTypes": [
    "admin.USER-UPDATE",
    "access.CUSTOM_REQUIRED_ACTION"
  ]
}
```
