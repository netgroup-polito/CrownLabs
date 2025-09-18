# Instance Inactive Termination Controller
CrownLabs includes a feature that automatically suspends or deletes Instances after a predefined period of user inactivity.
- `Persistent Instances` are only paused and can be restarted by the user from the web interface.
- `Non-persistent Instances` are permanently deleted.

Users can also receive email notifications before any action is taken, giving them the chance to prevent suspension or deletion.
This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- Setting `enable-instance-inactive-termination=true` enables inactivity monitoring. 
- Setting `enable-inactivity-notifications=true` enables email notifications, so users are warned several times before their instance is paused or deleted.

Each `Template` resource defines an `InactivityTimeout` field, which specifies the maximum period of inactivity allowed. For example:
- If `InactivityTimeout` is set to `10d` (10 days), CrownLabs will start sending notifications after 10 days without any access (via browser or SSH).
- If the user does not access the instance even after these warnings, the instance will be paused (if persistent) or deleted (if non-persistent).

By default, new Templates have `InactivityTimeout` set to `never`, meaning inactivity checks are disabled. Instances created from such templates will never be terminated due to inactivity.


# Instance Expiration Controller

CrownLabs includes a feature that automatically deletes Instances after they reach a maximum age (time passed from the creation).

Users can receive warning email notifications before the deletion take place.
This functionality is controlled by flags in the `crownlabs-instance-operator-automation` deployment manifest:
- Setting `enable-instance-expiration=true` enables the expiration monitoring. 
- Setting `enable-expiration-notifications=true` enables email notifications.

Each `Template` resource defines a `DeleteAfter` field, which specifies the maximum age allowed for Instances. For example:
- If `DeleteAfter` is set to `20d` (20 days), CrownLabs will send an email when the Instance is about to reach 20 days of age.
- When this timer expires, the Instance is automatically deleted without any possiiblity for the user to stop this action.

By default, new Templates have the `DeleteAfter` field set to `never`, meaning the deletion for expiration is disabled. Instances created from such templates do not have any maximum age contraint.
