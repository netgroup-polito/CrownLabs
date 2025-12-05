# Instance Inactive Termination Controller
CrownLabs includes a feature that automatically suspends or deletes Instances after a predefined period of user inactivity.

- **Persistent Instances** are paused and can be restarted by the user from the web interface.
- **Non-persistent Instances** are permanently deleted.

Users can receive email notifications before any action is taken, giving them the chance to prevent suspension or deletion. This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- `enable-instance-inactive-termination=true` enables inactivity monitoring.
- `enable-inactivity-notifications=true` enables email notifications, so users are warned several times before their instance is paused or deleted.

Each `Template` resource defines an `InactivityTimeout` field, which specifies the maximum period of inactivity allowed. For example:
- If `InactivityTimeout` is set to `10d` (10 days), CrownLabs will start sending notifications after 10 days without any access (via browser or SSH).
- If the user does not access the instance even after these warnings, the instance will be paused (if persistent) or deleted (if non-persistent).

By default, new Templates have `InactivityTimeout` set to `never`, meaning inactivity checks are disabled. Instances created from such templates will never be terminated due to inactivity.

## Namespace-level opt-out
You can prevent all instances in a namespace from being considered for inactivity-based suspension or deletion by adding the label `InstanceInactivityIgnoreNamespace: "true"` to the namespace. If this label is present and set to `true`, the inactivity controller will ignore all instances in that namespace, regardless of their template settings. If omitted or set to `false`, inactivity checks are enabled as usual.

## Additional details
- The number of warning notifications sent before action is taken is controlled by the `inactiveTerminationMaxNumberOfAlerts` Helm parameter, and can be overridden per-template using the `CustomNumberOfAlertsAnnotation` annotation.
- The time between notifications is controlled by the `inactiveTerminationNotificationInterval` Helm parameter.
- The controller uses Prometheus metrics (Nginx and SSH) to determine last activity.
- The following annotations and labels are used:
	- `AlertAnnotationNum`: Number of notifications sent to the tenant.
	- `LastNotificationTimestampAnnotation`: Timestamp of the last notification sent.
	- `LastRunningAnnotation`: Used to detect if the instance has been restarted after being paused.
	- `CustomNumberOfAlertsAnnotation`: Overrides the default max number of alerts for a specific template.

For more technical details, see the [Instance Automation Controller README](../pkg/instautoctrl/README.md).



# Instance Expiration Controller

CrownLabs includes a feature that automatically deletes Instances after they reach a maximum age (time passed from creation), as defined by the `DeleteAfter` field in the associated Template.

Users can receive warning email notifications before deletion takes place. This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- `enable-instance-expiration=true` enables expiration monitoring.
- `enable-expiration-notifications=true` enables email notifications.

Each `Template` resource defines a `DeleteAfter` field, which specifies the maximum age allowed for Instances. For example:
- If `DeleteAfter` is set to `20d` (20 days), CrownLabs will send an email when the Instance is about to reach 20 days of age.
- When this timer expires, the Instance is automatically deleted without any possibility for the user to stop this action.

By default, new Templates have the `DeleteAfter` field set to `never`, meaning expiration-based deletion is disabled. Instances created from such templates do not have any maximum age constraint.

## Namespace-level opt-out
If the namespace containing an instance has the label `ExpirationIgnoreNamespace: "true"`, expiration checks are also skipped for all instances in that namespace.

## Additional details
- The time between warning and deletion is controlled by the `expirationNotificationInterval` Helm parameter.
- The controller uses the annotation `ExpiringWarningNotificationTimestampAnnotation` to track if a warning has already been sent and when.
- The controller only deletes instances when the maximum age is reached, regardless of activity.

For more technical details, see the [Instance Automation Controller README](../pkg/instautoctrl/README.md).
