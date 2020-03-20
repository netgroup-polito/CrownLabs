# Students Creator
This Python script allows the creation of student accounts in Keycloak with their companion set of resources, namely Namespaces, Secrets and RoleBindings in Kubernetes.

## How does it work
This scripts creates the required resources associated to a user account, then sends an email to each new user to set the password for the account. In case the user already exists, this script adds a new binding for the given user to the new group.

This script must work in the same folder with the [namespace_template.yaml](namespace_template.yaml) [regcred_template.yaml](regcred_template.yaml) and the [role_binding_template.yaml](role_binding_template.yaml) and a CSV file containing the list of users that have to be created. 
The CSV file must follow the template such as [example.csv](example.csv).

To use this script in your cluster, you have to customize the following information:
- file [regcred_template.yaml](regcred_template.yaml): configure the secret in the `.dockerconfigjson` field;
- file [adduser.py](adduser.py): configure the URL of your OICD server in the `server_url` field.

## Dependencies
The following libraries must be present in order for this script to work:
- python-keycloak
- pandas
- jinja2

To install those dependencies you can run the following command:
````
 pip3 install <library-name>
````

## How run it
The name of CSV file is passed on the command line together with the username and password of the keycloak administrator:
````
 python3 adduser.py <csvfile> <username> <password>
````
