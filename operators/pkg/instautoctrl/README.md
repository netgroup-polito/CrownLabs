- [Instance Automation Controller](#instance-automation-controller)
  - [Instance Inactive Termination Controller](#instance-inactive-termination-controller)
    - [Detailed Behavior](#detailed-behavior)
    - [Inactivity detection](#inactivity-detection)
    - [Watch and Predicates for the reconciler](#watch-and-predicates-for-the-reconciler)
    - [Labels and Annotations](#labels-and-annotations)
  - [Instance Expiration Controller](#instance-expiration-controller)
    - [Detailed Behavior](#detailed-behavior-1)
    - [Watch and Predicates for the reconciler](#watch-and-predicates-for-the-reconciler-1)
    - [Labels and Annotations](#labels-and-annotations-1)
  - [Instance Termination Controller](#instance-termination-controller)
  - [Instance Submission Controller](#instance-submission-controller)
  - [Helm Chart](#helm-chart)

# Instance Automation Controller

The **Instance Automation Controller** (`instautoctrl` package) handles all automation tasks related to Instances, covering four main areas: instance inactivity management, instance expiration handling, instance termination processes, and instance submission workflows.

The package includes four different controllers:
- Instance Inactive Controller
- Instance Expiration Controller
- Instance Termination Controller
- Instance Submission Controller


## Instance Inactive Termination Controller
This controller monitors instances and automates actions based on their inactivity status and lifespan.
The controller understands if the Instance can be declared as Inactive and starts sending notifications to its tenant to inform them to access their Instance resources, otherwise they will be paused (if persistent) or deleted (if not persistent) after a specific period of time defined in the `Template` resource.
The template introduces a new field called `InactivityTimeout`, which specifies the period of time during which the Tenant must not access the Instance for it to be considered inactive and eligible for deletion.
This field is always available in the Template resource and, if omitted, it is set to `never` by default, meaning that Instances of that Template will be ignored by the controller.


### Detailed Behavior
The controller begins by retrieving all the active **Instances**.
For each instance, it determines whether it should be monitored or not.

Once it realizes that the instance should be monitored, the controller adds several annotations to it.
These include the `AlertAnnotationNum`, which tracks the number of notifications sent to inform the tenant that the instance has been idle for some time and will soon be stopped or deleted.
This number ranges from zero up to a maximum limit defined by `InstanceMaxNumberOfAlerts`, a custom parameter defined via Helm chart.
This value could be overwritten by the `CustomNumberOfAlertsAnnotation` annotation in the associated Template resource. Another annotation is the `LastActivityAnnotation`, which records the last time the user accessed the instance either via the frontend through the Ingress or via SSH (info available through the SSH bastion tracker).
When a new email notification is triggered, the controller adds the `LastNotificationTimestampAnnotation` to the instance, recording the timestamp of the last sent notification (if the feature is enabled). This annotation is used to determine whether the required interval has elapsed since the previous notification, thereby allowing a new alert to be sent if necessary.
The Helm Chart introduces the `notificationInterval` parameter, which defines the minimum time interval between two consecutive email notifications.
If the interval has passed, a new email is sent; otherwise, the notification is skipped.

Next, the controller checks whether the instance is inactive by comparing its last activity timestamp with the **InactivityTimeout** value specified in the Template. 
If the instance is found to be inactive (meaning the remaining time is zero or less), a series of notification emails are sent to the instance owner.
Once the number of alerts reaches a configurable threshold (either defined via `InstanceMaxNumberOfAlerts` or `CustomNumberOfAlertsAnnotation`), CrownLabs will take action by either stopping the instance if it is persistent, or deleting it if it is non-persistent.
On the other hand, if the instance is still active (the remaining time is greater than zero), the controller evaluates the remaining time and reschedules the inactivity check when it expires (a one-minute margin is added to the timer to be sure the timer is actually expired).

Finally, if the instance has been paused and the user restarts it, the `AlertAnnotationNum` annotation is reset and the `LastActivityAnnotation` annotation is updated.
The controller then evaluates the new remaining time, and the entire monitoring process begins again.
This mechanism relies on the `LastRunningAnnotation` annotation to detect if the instance has been restarted after being paused.


### Inactivity detection
The controller focuses on one point: understanding if the **Instance** is being used (and it should not be deleted) or it is not being used (and it should be deleted).
An Instance can be accessed by the Crownlabs Frontend or via SSH.
The controller uses **Prometheus** to do this check:
* It uses Nginx metrics to verify the last access to the Frontend
* It uses a custom metric (called **bastion_ssh_connections**) to monitor the SSH accesses. Read [here](../../README.md#bastion-ssh-tracker) for more info on how SSH connections are monitored.

Note: a single query on Prometheus cannot return more than **11000 data points**. In order to cover all the scenarios, a new parameter **queryStep** has been defined in the Helm Chart to modify the query resolution (query step), based on the **InactivityTimeout** selected.

After this check, the **LastActivityAnnotation** is updated with the most recent timestamp.
If the last access is above the max threshold (defined with the `inactivityTimeout` field in the **Template** resource), the Instance is declared as **inactive** and (if enabled) email notifications start to be sent at regular interval (**NotificationInterval** parameter in the Helm chart).
After the maximum time of notifications, the Instance is stopped.


### Watch and Predicates for the reconciler
The **InstanceInactiveTerminationReconciler** is set to watch and react to events related to the following resources in an efficient way:
* **Instances**: if an Instance has been stopped and the user restart is, the reconciler on that Instance must be triggered again to restart the monitoring process. There is a predicate filter (**instanceTriggered**) to let the reconciler reschedule the Instance.
* **Templates**: if the `inactivityTimeout` is set or modified in a template, the associated instances must be reconciled to recalculate the remaining time of the associated instances.
* **Namespaces**: if a `Namespace` is set to be monitored (`InstanceInactivityIgnoreNamespace != true`), all the Instance of that `Namespace` must be reconciled to evaluate the remaining time of the instance. There is a predicate filter (called **inactivityIgnoreNamespace**) to let the reconciler reschedule the Instance if a new `Namespace` has to be checked.


### Labels and Annotations
* **InstanceInactivityIgnoreNamespace**: `Namespace` annotation used to ignore the inactivity termination for all the Instances of the entire `Namespace`. Default value (if omitted) is `false`.
* **AlertAnnotationNum**: Instance annotation that stores the number of email notifications sent to the `Tenant`.
* **LastNotificationTimestampAnnotation**: Instance annotation that stores the timestamp of the last email notification sent to the `Tenant`.
* **LastRunningAnnotation**: Instance annotation that stores the previous value of the **Running** field of the Instance. It is used to check whether the `Instances` have been restarted after being paused.
* **CustomNumberOfAlertsAnnotation**: Template annotation that stores the override the default `InstanceMaxNumberOfAlerts` in the **InstanceInactiveTerminationReconciler** for a specific template.


## Instance Expiration Controller
This controller verifies whether the instance has exceeded its maximum lifespan, as defined by the **DeleteAfter** field in the associated **Template** resource.
If exceeded, the instance and its related resources are deleted.


### Detailed Behavior
The controller retrieves all the active Instances and fetches the related **Template** resource. Based on the `DeleteAfter` field of the Template, the maximum lifespan of each Instance is determined. 

When omitted, this value is set to never, meaning the Instance is not scheduled for termination. However, it can be configured with a time interval representing durations in minutes, hours, or days.

Once the instance lifespan expires, the controller sends a warning email to the tenant informing them that their Instance will be deleted soon.
The controller adds a new `ExpiringWarningNotificationAnnotation` annotation to the Instance to know that the first warning notification has been sent.
After the `notificationInterval` time has passed since the warning, the controller proceeds to delete the Instance and sends a second email to the tenant confirming that the Instance has been deleted.


### Watch and Predicates for the reconciler
The **InstanceExpirationReconciler** is set to watch and react to events related to the following resources in an efficient way:
* **Instances**: if an Instance has been stopped and the user restart is, the reconciler on that Instance must be triggered again to restart the monitoring process. There is a predicate filter (**instanceTriggered**) to let the reconciler reschedule the Instance.
* **Templates**: if the `deleteAfter` value is set or modified in a template, the associated instances must be reconciled to recalculate the remaining time of the associated instances. There is a predicate filter (**deleteAfterChanged**) to let the reconciler reschedule the Instance to update the new remaining time.
* **Namespaces**: if a `Namespace` is set to be monitored (`ExpirationIgnoreNamespace != true`), all the Instance of that `Namespace` must be reconciled to evaluate the remaining time of the instance. There is a predicate filter (called **expirationIgnoreNamespace**) to let the reconciler reschedule the Instance if a new `Namespace` has to be checked.


### Labels and Annotations
* **ExpirationIgnoreNamespace**: `Namespace` label used to ignore the expiration for all the Instances of the entire `Namespace`. Default value (if omitted) is `false`.
* **ExpiringWarningNotificationAnnotation**: Instance annotation that stores whether a warning notification has already been sent to the `Tenant` before the Instance expiration.


## Instance Termination Controller
This controller specifically focuses on instance termination in **exam scenarios**.
It first verifies whether the instance’s public endpoint is still responding by performing an HTTP check.
If the endpoint is found to be unreachable, the controller proceeds to initiate the termination process for that instance.


## Instance Submission Controller
This controller automates **exam submission** workflows by creating a ZIP archive of the instance’s persistent volume, which contains the VM disk.
Once the archive is created, it is uploaded to a configured submission endpoint.
This process is used during exams to collect student submissions in a reproducible and traceable way, ensuring consistency and accountability.


## Helm Chart

The Instance Automation Controller is deployed together with the the Instance Operator as a secondary deployment. The Helm chart for the Instance Operator has been updated to include the deployment of the new Instance Automation controller. 

Main controller parameters:

* **mailTemplateDir**: path to the directory containing the crownmail templates.
* **mailConfigDir**: path to the directory containing the crownmail configuration files.

Main automation parameters:

* **enableInstanceSubmission**: flag to enable the Instance Submission controller.
* **enableInstanceTermination**: flag to enable the Instance Termination controller.
* **enableInstanceInactiveTermination**: flag to enable the Instance Inactive Termination controller.
* **enableInstanceExpiration**: flag to enable the Instance Expiration controller.
* **inactiveTerminationMaxNumberOfAlerts**: maximum number of email notifications to send to the Tenant before deleting/pausing the Instance.
* **enableInactivityNotifications**: flag to enable the notification for the Instance Inactive Termination controller.
* **enableExpirationNotifications**: flag to enable the notification for the Instance Expiration Termination controller.
* **inactiveTerminationNotificationInterval**: time interval between two consecutive email notifications.
* **expirationNotificationInterval**: time interval between the warning notification and the Instance deletion.

Main monitoring parameters:

* **prometheusURL**: URL of the Prometheus service in the cluster.
* **queryNginxAvailable**: query to verify if the external ingress is available and is correctly collecting metrics.
* **queryBastionSSHAvailable**: query to verify if the custom SSH bastion tracker is available and is correctly collecting metrics.
* **queryWebSSHAvailable**: query to verify if the WebSSH (SSH through a new browser terminal) metrics are available.
* **queryNginxData**: query to retrieve info about an Instance access through frontend.
* **queryBastionSSHData**: query to retrieve info about an Instance access through SSH.
* **queryWebSSHData**: query to retrieve info about an Instance access through WebSSH.
* **queryStep**: step to use in the Prometheus query to retrieve data.