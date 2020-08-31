# Creating Courses, Labs and student accounts
This Python script allows the automatic creation of *CrownLabs* courses, including the different laboratories and the tenant (i.e. student and professor) accounts. In particular, tenants are characterized both by the actual accounts in the OIDC server (i.e. Keycloak), as well as their companion set of resources in Kubernetes.

## Overview
This script processes different CSV files to obtain the information regarding the resources characteristics of the resources to be created. The files must be compliant with the templates provided in the [csv-examples](csv-examples) folder:

* [courses.csv](csv-examples/courses.csv): enumerates the list of courses to be created by the script;
* [laboratories.csv](csv-examples/laboratories.csv): enumerates the list of laboratories to be created by the script;
* [students.csv](csv-examples/students.csv): enumerates the list of student accounts to be created by the script;
* [teachers.csv](csv-examples/teachers.csv): enumerates the list of professor accounts (i.e. with additional privileges) to be created by the script.

Upon the creation of a new tenant account, the system automatically sends a welcome email to the tenant. In particular, the email contains a confirmation link that needs to be accessed to complete the registration and setup a new password for the account.

The script is designed to be idempodent, i.e. it can be executed multiple times with the same inputs and it will always produce the same results. Additionally, the modifications are incremental, e.g. it is possible to introduce new laboratories or allow tenants to access additional courses even after the initial creation.

The script depends upon a series of Kubernetes resource templates stored within the [templates](templates) folder.
Before executing the script, it is possible to customize the following files:
- [setup-courses.py](setup-courses.py): optional, to configure the URL of a different OICD server (`server_url` field).

## Dependencies
The following libraries must be present in order for this script to work:
- `python-keycloak`
- `pandas`
- `jinja2`
- `requests`
- `secrets`

To install those dependencies you can run the following command:
````
 pip3 install -r requirements.txt
````

## Usage

```
usage: setup-courses.py [-h] [-c <courses.csv>] [-l <laboratories.csv>]
                        [-t <teachers.csv>] [-s <students.csv>]
                        keycloak_user keycloak_pass nextcloud_user nextcloud_pass
```

#### Positional arguments:

* `keycloak_user`: The admin username for the OIDC server;
* `keycloak_pass`: The admin password for the OIDC server;
* `nextcloud_user`: The admin username for Nextcloud server;
* `nextcloud_pass`: The admin password for Nextcloud server;

#### Optional arguments:

* `-h, --help`: Show this help message and exit;
* `-c <courses.csv>, --courses <courses.csv>`: The CSV file containing the courses to be created
* `-l <laboratories.csv>, --laboratories <laboratories.csv>`: The CSV file containing the list of laboratories to be created;
* `-t <teachers.csv>, --teachers <teachers.csv>`: The CSV file containing the professor accounts to be created;
* `-s <students.csv>, --students <students.csv>`: The CSV file containing the student accounts to be created;

**Note:** the optional arguments specifying the input files can be provided both all together or during multiple runs of the script. The only constraint relates to the courses: courses must be either already present before creating the other resources or the `--courses <courses.csv>` parameter needs to be specified.
