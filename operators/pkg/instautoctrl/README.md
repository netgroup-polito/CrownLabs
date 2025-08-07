# Instance Automation Controller

The **Instance Automation Controller** (`instautoctrl` package) is part of the automation framework developed for the *Cloud Programming* course, Politecnico di Torino, 2025 edition.  
It implements backend logic to manage the lifecycle of instances declared as **Inactive**.

The package includes four controllers:

- [Instance Inactive Termination](#instance-inactive-termination-controller)
- [Instance Termination](#instance-termination-controller)
- [Instance Submission](#instance-submission-controller)
- [Instance Expiration](#instance-expiration-controller)


## Instance Inactive Termination Controller
This controller monitors instances and automates actions based on their inactivity status and lifespan. The controller understands if the Instance can be declared as Inactive and starts sending notifications to its tenant to inform them to access their Instance resources, otherwise they will be paused (if persistent) or deleted (if not persistent) after a specific period of time defined in the `Template` resource.


### Detailed Behavior
The controller begins by retrieving all the active **Instances**. For each instance, it determines whether it should be monitored or not. This decision is influenced by a special label that can be added to the associated **Namespace** resource, called `InstanceInactivityIgnoreNamespace`. If the Namespace carries this label, the controller will ignore the instance, allowing it to remain inactive for an extended period without being stopped or deleted.

Once it is established that the instance should be monitored, the controller adds several annotations to it. These include the `AlertAnnotationNum`, which tracks the number of notifications sent to inform the tenant that the instance has been idle for some time and will soon be stopped or deleted. This number ranges from zero up to a maximum limit defined by `InstanceMaxNumberOfAlerts`. This value could be overwritten by the `CustomNumberOfAlertsAnnotation`. Another annotation is the `LastActivityAnnotation`, which records the last time the user accessed the instance either via the frontend through the Ingress or via SSH (through the SSH bastion tracker). The controller also adds the `LastNotificationTimestampAnnotation`, indicating the timestamp when the last notification was sent. This timestamp helps to determine if enough time has passed since the previous notification, allowing a new alert to be sent if needed.

Next, the controller checks whether the instance is inactive by comparing its last activity timestamp with the **InactivityTimeout** value specified in the Template. If the instance is found to be inactive (meaning the remaining time is zero or less), a series of notification emails are sent to the instance owner. Once the number of alerts reaches a configurable threshold, CrownLabs will take action by either stopping the instance if it is persistent, or deleting it if it is non-persistent. On the other hand, if the instance is still active (the remaining time is greater than zero), the controller reschedules the inactivity check for a future time.

Finally, if the instance has been paused and the user restarts it, the `AlertAnnotationNum` is reset. The controller then evaluates the new remaining time, and the entire monitoring process begins again. This mechanism relies on the `LastRunningAnnotation` annotation to detect if the instance has been restarted after being paused.


### How does the check is performed?
The controller focuses on one point: understanding if the **Instance** is being used (and it should not be deleted) or it is not being used (and it should be deleted).
An Instance can be accessed by the Crownlabs Frontend or via SSH.
The controller uses **Prometheus** to do this check:
* It uses Nginx metrics to verify the last access to the Frontend
* It uses a custom metric (called **bastion_ssh_connections**) to monitor the SSH accesses.

After this check, the **LastActivityAnnotation** is updated with the most recent timestamp. If the last access is above the max threshold (defined with the `inactivityTimeout` field in the **Template** resource), the Instance is declared as **inactive** and (if enabled) email notifications start to be sent at regular interval (**NotificationInterval** parameter).
After the maximum time of notifications, the Instance is stopped.


### Watch and Predicates for the reconciler
The **InstanceInactiveTerminationReconciler** is set to watch and react to events related to the following resources:
* **Instances**: if an Instance has been stopped and the user restart is, the reconciler on that Instance must be triggered again to restart the monitoring process. There is a predicate filter (**instanceTriggered**) to let the reconciler reschedule the Instance.
* **Templates**: if the `inactivityTimeout` is set or modified in a template, the associated instances must be reconciled to recalculate the remaining time of the associated instances.
* **Namespaces**: if a `Namespace` is set to be monitored (add `InstanceInactivityIgnoreNamespace` label), all the Instance of that `Namespace` must be reconciled to evaluate the remaining time of the instance. There is a predicate filter (called **inactivityIgnoreNamespace**) to let the reconciler reschedule the Instance if a new `Namespace` has to be checked.


### Labels and Annotations
* **InstanceInactivityIgnoreNamespace**: label added to the `Namespace` to ignore the inactivity termination for the Instances in that `Namespace`. 
* **AlertAnnotationNum**: annotaion to check the number of email notifications already sent to the `Tenant`.
* **LastNotificationTimestampAnnotation**: annotation to check the timestamp of the last email notification sent to the `Tenant`.
* **LastRunningAnnotation**: previous value of the **Running** field of the Instance. It is used to check whether the `Instances` have been restarted after being paused.
* **CustomNumberOfAlertsAnnotation**: override the default `InstanceMaxNumberOfAlerts` in the **InstanceInactiveTerminationReconciler** for a specific template.
* **LastRunningAnnotation**:  previous value of the `Running` field of the Instance. It is used to check wheather a persistent Instance was stopped and now has been started again.


## Instance Termination Controller
This controller specifically focuses on instance termination in **exam scenarios**. It first verifies whether the instance’s public endpoint is still responding by performing an HTTP check. If the endpoint is found to be unreachable, the controller proceeds to initiate the termination process for that instance.


## Instance Submission Controller
This controller automates **exam submission** workflows by creating a ZIP archive of the instance’s persistent volume, which contains the VM disk. Once the archive is created, it is uploaded to a configured submission endpoint. This process is used during exams to collect student submissions in a reproducible and traceable way, ensuring consistency and accountability.


## Instance Expiration Controller
This controller is a replacement for the old `delete-stale-instance` python script. It verifies whether the instance has exceeded its maximum lifespan, as defined by the **DeleteAfter** field in the associated **Template** resource. If exceeded, the instance and its related resources are deleted.


### Detailed Behavior
When the controller starts, it retrieves all the **Instances** and, for each one, it fetches the related **Template** resource. Inside this Template, there is a field called `DeleteAfter` which specifies the maximum lifespan of that Instance. By default, this value is set to never, meaning the Instance is not scheduled for termination. However, it can be configured with a time interval representing durations in minutes, hours, or days.

Using the `DeleteAfter` value, the controller calculates the remaining lifespan of the Instance. Once this lifespan expires, the controller sends an email notification to the Instance owner (tenant) informing them that their Instance will be deleted. After a predefined extra waiting period, the controller proceeds to delete the Instance. Finally, it sends a second email to the tenant confirming that the Instance has reached its maximum lifespan and has been deleted.


### Watch and Predicates for the reconciler
The **InstanceExpirationReconciler** is set to watch and react to events related to the following resources:
* **Instances**: if an Instance has been stopped and the user restart is, the reconciler on that Instance must be triggered again to restart the monitoring process. There is a predicate filter (**instanceTriggered**) to let the reconciler reschedule the Instance.
* **Templates**: if the `deleteAfter` value is set or modified in a template, the associated instances must be reconciled to recalculate the remaining time of the associated instances. There is a predicate filter (**deleteAfterChanged**) to let the reconciler reschedule the Instance to update the new remaining time.
