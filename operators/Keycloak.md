# Keycloak Integration

The Operator is able to interact with Keycloak.
It has been developed using Keycloak v26.2.4, but a compatibility module is available to support older versions.

If you want to completely disable the Keycloak integration, you can set the `--keycloak-enabled` arg to `false` in the Operator configuration.

## Operator client

The Operator requires a dedicated Keycloak client to interact with the Keycloak server.

⚠️ **The client must be different from the one used to sign in the users (frontend and k8s API server).**
Otherwise, security issues may arise.

The Client must be configured with _Service Account roles_ and the following roles must be assigned to it:

- `realm-management / manage-users`
- `realm-management / query-clients`
- `realm-management / manage-clients`
- `realm-management / manage-events`

In some versions of Keycloak, the `view-events` and `view-clients` roles may also be required.

You can disable the _Standard flow_ for the client, as it is not used by the Operator.

## Webhooks

The Operator can receive events from Keycloak through webhooks.
These events are used to determine whether a user has been confirmed, and trigger the creation of the corresponding Kubernetes resources.

### Keycloak-side configuration

The events are sent by Keycloak using an external plugin, [keycloak-events](https://github.com/p2-inc/keycloak-events), that **must** be installed in the Keycloak server _before_ installing the Operator.
You can use the image `quay.io/phasetwo/phasetwo-keycloak:26.2.4` to run Keycloak with the plugin already installed.

Then, you have to enable the plugin in the Keycloak realm used by the Operator.
To do this, go to the Keycloak admin console, and navigate to **Realm Settings** -> **Events**, then add the `ext-event-webhook` provider.

### Operator-side configuration

If you have installed the Operator using the Helm chart, the webhook is then automatically configured.

Otherwise, you can configure it manually performing the following `POST` request to the Keycloak server:

_Endpoint_: `http://keycloak-endpoint.example.com/realms/REALM_NAME/webhooks`

_Body_:
```json
{
  "enabled": "true",
  "url": "http://operator-endpoint.example.com:8082/tenant-webhook",
  "eventTypes": [
    "admin.USER-UPDATE",
    "access.CUSTOM_REQUIRED_ACTION"
  ]
}
```

Pay attention to replace `keycloak-endpoint.example.com`, `operator-endpoint.example.com`, `REALM_NAME` with the actual values of your Keycloak and Operator endpoints and the name of the Keycloak realm used by the Operator.
Port `8082` is the default port used by the Operator to receive webhooks, but it can be changed in the Operator configuration.

The request must be authenticated using an account able to manage the Keycloak events.

## Compatibility mode

If you are using an older version of Keycloak, you can enable the compatibility mode in the Operator.
To do this, you can set the `--keycloak-compatibility-mode` arg to `true` in the Operator configuration.
