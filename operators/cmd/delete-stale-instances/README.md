
# Delete Stale Instances

## About the application

- The application deletes all active instances that are expired.
- The expiration threshold is retrieved from corresponding template (i.e, the `deleteAfter` field). Therefore who creates the instance is not able to modify (for now) the expiration time, since it is inherited from the template.
- You can see the expiration threshold definition in the [template](../../deploy/crds/crownlabs.polito.it_templates.yaml) specification.
- The expiration time has the following standard format `[0-9]+[mhd]` and represents the limit time for which an instance is allowed to run. Additionally, it can be set to `never`, to prevent the deletion of the corresponding instances.
- The expiration time is calculated from the creation timestamp of the instance, and the instance is deleted when the difference between the current timestamp and the creation timestamp exceeds the expiration threshold.
- The application is run through a cronjob.
