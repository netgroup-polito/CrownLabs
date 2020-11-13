====
## ABOUT THE APPLICATION 

- The application deletes all active instances that are expired. 
- The expiration threshold is retrieved from correspondind template
- The expiration time has the following standard format:[0-9]+[mhd] and
rappresents the limit time for which an instance is allowed to run. 
- The application is run through a cronjob, every 15 minutes.

## HOW TO RUN

To create the docker image run:
- cd inside /operators path
- docker build -f ./build/delete-stale-instances/Dockerfile -t yourDockerRepo:version .
To run the application do the following steps:
- kubectl create -f k8s-manifest.yaml
- kubectl create -f k8s-cluster-role.yaml
To see the cronjob running: 
- kubectl get cronjobs --namespace crownlabs-delete-stale-instances-space
To see instances running:
- kubectl get instances --namespace instance_namespace