# Creating Courses, Labs and Student/Professor accounts
This Python script allows the automatic **creation** of *CrownLabs* **courses**, including the different **laboratories** and the accounts for **students and professors**.
In particular, the script takes care of creating the proper accounts in the OIDC server (i.e. Keycloak) as well as the companion set of resources in Kubernetes (e.g., namespaces).

This script can be used also to **update** an existing configuration (more details below).

## Overview
This script processes different CSV files to obtain the information about the resources that have to be created.
Files must be compliant with the templates provided in the [csv-examples](csv-examples) folder:

* [courses.csv](csv-examples/courses.csv): enumerates the list of courses (e.g., "Computer Networks") to be created by the script;
* [laboratories.csv](csv-examples/laboratories.csv): enumerates the list of laboratories (e.g., "Lab1 - Traffic analysis", "Lab 2 - Routing"), to be created by the script;
* [students.csv](csv-examples/students.csv): enumerates the list of student accounts to be created by the script;
* [teachers.csv](csv-examples/teachers.csv): enumerates the list of professor accounts (i.e. with additional privileges) to be created by the script.

The script can either create **all the resources at once** (courses, lab, users), or a **single type of resources** (e.g., users).
The second option is particularly useful when some data has to be updated, such as adding some new students to the current list of enrolled students. 
In fact, the script is designed to be idempodent, i.e. it can be executed multiple times with the same inputs and it will always produce the same results.
Additionally, modifications are incremental, e.g. it is possible to introduce new laboratories, allow tenants to access additional courses, or add new students to an existing course, at any time.
Existing information is kept, while new information is added to the database.

The script depends upon a series of Kubernetes resource templates stored within the [templates](templates) folder.
Before executing the script, it is possible to customize the URL of the OICD server (`server_url` field in [setup-courses.py](setup-courses.py)).

Upon the creation of a new tenant account, the system automatically sends a welcome email to the tenant. In particular, the email contains a confirmation link that needs to be accessed to complete the registration and setup a new password for the account.


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

## Requirements
This script needs to interact with your Kubernetes setup.
Hence, it requires `kubectl` to be available, and configured with the correct `kubeconfig` (bound to the `cluster-admin` role) to interact with the Kubernetes cluster hosting CrownLabs.
This privileged access to your Kubernetes cluster can usually be achieved if you log-in in the master machine of your Kubernetes domain.

## Usage

```
usage: python3 setup-courses.py [-h] [-c <courses.csv>] [-l <laboratories.csv>]
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
