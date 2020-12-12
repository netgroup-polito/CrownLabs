
# Delete Stale Instances

## ABOUT THE APPLICATION

- The application deletes all active instances that are expired.
- The expiration threshold is retrieved from corresponding template. This value is set when the template is created in the deleteAfter field or, if not specified, it assumes the default value of 7 days. Therefore who creates the instance is not able to modify ( for now) the expiration time, but it inherits it from the template. 
- You ca see the expiration threshold in the [template](../../deploy/crds/crownlabs.polito.it_templates.yaml) definition.
- The expiration time has the following standard format `[0-9]+[mhd]` and
represents the limit time for which an instance is allowed to run.
- The expiration time is calculated from the creation timestamp of the instance, indeed the instance is deleted when the difference between the current timestamp and the creation timestamp exceeds the expiration threshold.
- The application is run through a cronjob, every 15 minutes.

## HOW TO RUN

To create the docker image run:
- cd inside /operators path
- docker build -f ./build/delete-stale-instances/Dockerfile -t yourDockerRepo:version

To run the application do the following steps:
- kubectl create -f k8s-manifest.yaml
- kubectl create -f k8s-cluster-role.yaml

To see the cronjob running:
- kubectl get cronjobs --namespace crownlabs-delete-stale-instances
