# Instance Inactive Termination Controller
CrownLabs includes a feature that automatically suspends or deletes Instances after a predefined period of user inactivity.

- **Persistent Instances** are paused and can be restarted by the user from the web interface.
- **Non-persistent Instances** are permanently deleted.

Users can receive email notifications before any action is taken, giving them the chance to prevent suspension or deletion. This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- `enable-instance-inactive-termination=true` (`enableInstanceInactiveTermination` value in the Helm chart) enables inactivity monitoring.
- `enable-inactivity-notifications=true` (`enableInactivityNotifications` value in the Helm chart) enables email notifications, so users are warned several times before their instance is paused or deleted.

Each `Template` resource defines an `InactivityTimeout` field, which specifies the maximum period of inactivity allowed. For example:
- If `InactivityTimeout` is set to `10d` (10 days), CrownLabs will start sending notifications after 10 days without any access (via browser or SSH).
- If the user does not access the instance even after these warnings, the instance will be paused (if persistent) or deleted (if non-persistent).

By default, new Templates have `InactivityTimeout` set to `never`, meaning inactivity checks are disabled. Instances created from such templates will never be terminated due to inactivity.

## Namespace-level opt-out
You can prevent all instances in a namespace from being considered for inactivity-based suspension or deletion by adding the label `crownlabs.polito.it/instance-inactivity-ignore: "true"` to the namespace. If this label is present and set to `true`, the inactivity controller will ignore all instances in that namespace, regardless of their template settings. If omitted or set to `false`, inactivity checks are enabled as usual.

## Additional details
- The number of warning notifications sent before action is taken is controlled by the `inactiveTerminationMaxNumberOfAlerts` Helm parameter, and can be overridden per-template using the `crownlabs.polito.it/custom-number-alerts` annotation.
- The time between notifications is controlled by the `inactiveTerminationNotificationInterval` Helm parameter.
- The controller uses Prometheus metrics (Nginx and SSH) to determine last activity.
- The following annotations are used:
	- `crownlabs.polito.it/number-alerts-sent`: Number of notifications sent to the tenant.
	- `crownlabs.polito.it/last-notification-timestamp`: Timestamp of the last notification sent.
	- `crownlabs.polito.it/last-running`: Used to detect if the instance has been restarted after being paused.
	- `crownlabs.polito.it/custom-number-alerts`: Overrides the default max number of alerts for a specific template.

## Phase 2: Inactivity Destruction After Powered Off (DestroyAfterInactivity)
A second timeout, `DestroyAfterInactivity`, can be specified in the `Template` resource. 
This applies exclusively to persistent instances left powered off for a specific time.
- If `DestroyAfterInactivity` is set to `10d`, CrownLabs will track how long the instance has been powered off.
- As soon as this limit is reached, new warning notifications will be sent to the user.
- If the user still doesn't turn the instance back on, the persistent instance will be permanently **destroyed** to completely free up resources (including the persistent volume).

Annotations added for this feature:
- `crownlabs.polito.it/last-powered-off-timestamp`: Timestamp of the last time the instance was powered off.
- `crownlabs.polito.it/destruction-alerts-sent`: Number of mail sent to the tenant to inform that the instance will be destroyed.
- `crownlabs.polito.it/last-destruction-notification-timestamp`: Timestamp of the last notification sent to the tenant to inform that the instance will be destroyed.

For more technical details, see the [Instance Automation Controller README](../pkg/instautoctrl/README.md).



# Instance Expiration Controller

CrownLabs includes a feature that automatically deletes Instances after they reach a maximum age (time passed from creation), as defined by the `DeleteAfter` field in the associated Template.

Users can receive warning email notifications before deletion takes place. This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- `enable-instance-expiration=true` (`enableInstanceExpiration` value in the Helm chart) enables expiration monitoring.
- `enable-expiration-notifications=true` (`enableExpirationNotifications` value in the Helm chart) enables email notifications.

Each `Template` resource defines a `DeleteAfter` field, which specifies the maximum age allowed for Instances. For example:
- If `DeleteAfter` is set to `20d` (20 days), CrownLabs will send an email when the Instance is about to reach 20 days of age.
- When this timer expires, the Instance is automatically deleted without any possibility for the user to stop this action.

By default, new Templates have the `DeleteAfter` field set to `never`, meaning expiration-based deletion is disabled. Instances created from such templates do not have any maximum age constraint.

## Namespace-level opt-out
If the namespace containing an instance has the label `crownlabs.polito.it/expiration-ignore: "true"`, expiration checks are also skipped for all instances in that namespace.

## Additional details
- The time between warning and deletion is controlled by the `expirationNotificationInterval` Helm parameter.
- The controller uses the annotation `crownlabs.polito.it/expiring-warning-notification-timestamp` to track if a warning has already been sent and when.
- The controller only deletes instances when the maximum age is reached, regardless of activity.

For more technical details, see the [Instance Automation Controller README](../pkg/instautoctrl/README.md).
