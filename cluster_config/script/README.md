# Students Creator
This script allows you to automatically create student accounts  and their Namespaces, Secrets and RoleBindings in Kubernetes.

## How does it work

This script must work in the same folder with the [namespace_template.yaml](namespace_template.yaml) [regcred_template.yaml](regcred_template.yaml)and the [role_binding_template.yaml](role_binding_template.yaml) and the csv file. 
Then send emails to new accounts to create the new password. The csv file must follow the template like that of [example.csv](example.csv)
## Dependencies
Before use it you must have those libraries:
* python-keycloak
* pandas
* jinja2

For install those dependencies you can run the following command
````
 pip3 install <library-name>
````
## How run it
The name of CSV file is passed on the command line together with the username and password of the keycloak administrator.
````
 python3 script.py <csv file> <username> <Password>
````
