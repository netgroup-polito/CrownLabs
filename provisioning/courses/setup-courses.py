import argparse
import base64
import json
import os
import sys
import jinja2 as Template
import keycloak as kc
import pandas as pd
import requests
import secrets


class KeycloakHandler:

    class ClientIDNotFound(Exception):
        pass

    class GroupNotFound(Exception):
        pass

    class ClientRoleNotFound(Exception):
        pass

    def __init__(self, admin_user, admin_pass, client_id):
        self.keycloak_admin = kc.KeycloakAdmin(
            server_url="https://auth.crownlabs.polito.it/auth/",
            username=admin_user,
            password=admin_pass,
            user_realm_name="master",
            realm_name="crownlabs",
            verify=True)

        self.client_id = self.keycloak_admin.get_client_id(client_id)
        if not self.client_id:
            raise KeycloakHandler.ClientIDNotFound("Client ID '{}' not found in Keycloak".format(client_id))

        self.groups = dict()
        self.client_roles = dict()

    def get_user(self, username):
        _users = self.keycloak_admin.get_users({'username': username})
        return None if len(_users) == 0 else _users[0]

    def create_new_user(self, email, username, first_name, last_name, namespace):
        _user_data = {
            'email': email,
            'username': username,
            'firstName': first_name,
            'lastName': last_name,
            'enabled': True,
            'emailVerified': False,
            'attributes': {'namespace': namespace, }
        }

        _user_id = self.keycloak_admin.create_user(_user_data)
        self.keycloak_admin.send_update_account(
            user_id=_user_id,
            payload='["UPDATE_PASSWORD","VERIFY_EMAIL"]',
            lifespan=KeycloakHandler.__get_email_lifespan())

        _user_data['id'] = _user_id
        return _user_data

    def add_namespace_attribute(self, user, namespace):
        _username = user.get("username")

        _attributes = user.get('attributes', {})
        _namespace = _attributes.get("namespace")

        if namespace != _namespace:
            _user_id = user.get('id')
            _attributes['namespace'] = namespace
            self.keycloak_admin.update_user(user_id=_user_id, payload={'attributes': _attributes})

    def add_course_attribute(self, user, course_code):
        _user_id = user.get('id')
        _attributes = user.get('attributes', {})
        _courses = _attributes.get('courses', [])
        _attributes['courses'] = list(set(_courses + [course_code, ]))
        self.keycloak_admin.update_user(user_id=_user_id, payload={'attributes': _attributes})

    def add_user_to_group(self, user, group_name):
        _user_id = user.get('id')

        self.keycloak_admin.group_user_add(user_id=_user_id, group_id=self.get_group(group_name))
        self.keycloak_admin.assign_client_role(user_id=_user_id, client_id=self.client_id,
                                               roles=self.get_client_role(group_name))

    def get_group(self, group_name):
        if not group_name in self.groups:
            _output = self.keycloak_admin.get_group_by_path("/{}".format(group_name))
            self.groups[group_name] = None if _output is None else _output["id"]

        if self.groups[group_name] is None:
            raise KeycloakHandler.GroupNotFound("Group {} does not exist".format(group_name))

        return self.groups[group_name]

    def create_group(self, group_name):
        self.keycloak_admin.create_group(payload={'name': group_name}, skip_exists=True)

    def get_client_role(self, role_name):
        if not role_name in self.client_roles:
            self.client_roles[role_name] = self.keycloak_admin.get_client_role(self.client_id, role_name)

        if self.client_roles[role_name] is None:
            raise KeycloakHandler.ClientRoleNotFound("Client role {} does not exist".format(role_name))

        return self.client_roles[role_name]

    def create_client_role(self, role_name):
        _output = self.keycloak_admin.create_client_role(
            client_role_id=self.client_id, skip_exists=True,
            payload={'name': role_name, 'clientRole': True})

    @staticmethod
    def __get_email_lifespan():
        return 3600 * 24 * 30  # 30 days


class NextcloudHandler:

    class UserCreationFailed(Exception):
        pass

    def __init__(self, base_url, admin_user, admin_pass):
        self.base_url = base_url
        self.admin_user = admin_user
        self.admin_pass = admin_pass

    # https://docs.nextcloud.com/server/15/admin_manual/configuration_user/instruction_set_for_users.html#add-a-new-user
    def create_user(self, email, username, first_name, last_name):
        _url = self._build_nextcloud_url("users")
        _headers = self._build_nextcloud_headers()
        _user_data = {
            "userid": NextcloudHandler.__get_nextcloud_username(username),
            "password": NextcloudHandler.__generate_password(),
            "displayName": "{} {}".format(first_name, last_name),
        }

        _response = requests.post(url=_url, headers=_headers, data=_user_data)
        return NextcloudHandler.__process_response(_response, {
            100: lambda : (True, (_user_data["userid"], _user_data["password"])),
            102: lambda : (False, None),
        })

    # https://docs.nextcloud.com/server/15/admin_manual/configuration_user/instruction_set_for_users.html#edit-data-of-a-single-user
    def update_password(self, username):
        _userid = NextcloudHandler.__get_nextcloud_username(username)
        _url = self._build_nextcloud_url("users/{}".format(_userid))
        _headers = self._build_nextcloud_headers()
        _user_data = {
            "key": "password",
            "value": NextcloudHandler.__generate_password(),
        }

        _response = requests.put(url=_url, headers=_headers, data=_user_data)
        return NextcloudHandler.__process_response(_response, {
            100: lambda : (_userid, _user_data["value"]),
        })

    def _build_nextcloud_url(self, path):
        return "{}/ocs/v1.php/cloud/{}?format=json".format(self.base_url, path)

    def _build_nextcloud_headers(self):
        _credentials = "{}:{}".format(self.admin_user, self.admin_pass).encode("ascii")
        _auth = base64.b64encode(_credentials).decode("ascii")
        return {
            "Authorization": "Basic {}".format(_auth),
            "OCS-APIRequest": "true",
        }

    @staticmethod
    def __process_response(response, processors):
        if response.status_code == 200:
            _response_meta = json.loads(response.text)['ocs']['meta']
            _status = _response_meta['statuscode']

            if _status in processors.keys():
                return processors[_status]()
            raise NextcloudHandler.UserCreationFailed("{}: {}".format(_status, _response_meta["message"]))

        raise NextcloudHandler.UserCreationFailed("HTTP Status Code: {}".format(response.status_code))


    @staticmethod
    def __get_nextcloud_username(course_code):
        return "keycloak-{}".format(course_code)

    @staticmethod
    def __generate_password():
        return secrets.token_urlsafe()


class KubernetesHandler:
    class FatalError(Exception):
        pass

    @staticmethod
    def check_kubectl_cluster_admin():
        if os.system("kubectl auth can-i '*' '*' -A | grep --quiet 'yes'") != 0:
            raise KubernetesHandler.FatalError("Kubectl not found or not enough privileges to interact with the cluster")


    @staticmethod
    def add_resource_label(resource_type, resource_name, label_key, label_value):
        os.system(("kubectl label --overwrite {0} {1} {2}={3} | "
                   "sed 's/^/*** k8s: /g' | sed 's/not //g' | sed 's/$/ as {2}={3}/g'")
                  .format(resource_type, resource_name, label_key, label_value))

    @staticmethod
    def check_existence(resource_type, namespace_name, resource_name):
        return 0 == os.system("kubectl get {0} -n {1} {2} >/dev/null 2>/dev/null"
                              .format(resource_type, namespace_name, resource_name))


class KubernetesTemplateHandler:
    def __init__(self, template_name, template_base_path="./"):
        _template_loader = Template.FileSystemLoader(searchpath=template_base_path)
        _template_env = Template.Environment(loader=_template_loader)
        self.template = _template_env.get_template(template_name)

    def apply_template(self, **kwargs):
        _resource = self.template.render(**kwargs)
        os.system("echo '{}' | kubectl apply -f - | sed 's/^/*** k8s: /g'".format(_resource))


class Course:
    def __init__(self, record):
        self.course_code = record["Course (code)"].lower()
        self.course_name = record["Name"]
        self.namespace = Course.get_namespace_name(self.course_code)
        self.group_name = Course.get_group_name(self.course_code)
        self.group_name_admin = Course.get_group_name_admin(self.course_code)
        sys.stdout.write("Processing course {} ({})\n".format(self.course_code, self.course_name))

    def keycloak_configuration(self, keycloak_handler):
        sys.stdout.write("* Creating group/role {} in Keycloak\n".format(self.group_name))
        keycloak_handler.create_group(self.group_name)
        keycloak_handler.create_group(self.group_name_admin)
        keycloak_handler.create_client_role(self.group_name)
        keycloak_handler.create_client_role(self.group_name_admin)

    def kubernetes_configuration(self, k8s_templates):
        sys.stdout.write("* Creating Kubernetes resources for course {}\n".format(self.course_code))
        k8s_templates["namespace"].apply_template(
            namespace_name=self.namespace,
            rendered_name="{}".format(self.course_name).replace(" ", "_"),
            namespace_type=Course.get_namespace_type())
        k8s_templates["rolebinding-course"].apply_template(
            namespace_name=self.namespace,
            course_group=Course.get_group_name(self.course_code))
        k8s_templates["rolebinding-courseadmin"].apply_template(
            namespace_name=self.namespace,
            course_group_admin=Course.get_group_name_admin(self.course_code))
        k8s_templates["clusterrolebinding-courseadmin"].apply_template(
            course_group_admin=Course.get_group_name_admin(self.course_code))

    @staticmethod
    def get_namespace_type():
        return "course"

    @staticmethod
    def get_namespace_name(course_code):
        return "course-{}".format(course_code)

    @staticmethod
    def get_group_name(course_code):
        return "course-{}".format(course_code)

    @staticmethod
    def get_group_name_admin(course_code):
        return "course-{}-admin".format(course_code)


class Tenant:
    def __init__(self, record, is_admin=False):
        self.username = record['Username']
        self.email = record['Email']
        self.last_name = record['Last name'].title()
        self.first_name = record['First name'].title()
        self.course_code = record['Course (code)'].lower()
        self.namespace = Tenant.get_namespace_name(self.username)
        self.group_name = Tenant.get_group_name(self.username)
        self.is_admin = is_admin

        sys.stdout.write("Processing tenant {} ({} {}) - Course: {}{}\n"
                         .format(self.username, self.first_name, self.last_name,
                                 self.course_code, " - Admin" if self.is_admin else ""))

    def keycloak_configuration(self, keycloak_handler):
        self.user = keycloak_handler.get_user(self.username)
        self._create_user(keycloak_handler)
        # Disable for the moment, since it creates problems with the GUI
        # self._assign_self_group(keycloak_handler)
        self._assign_course_group(keycloak_handler)

    def nextcloud_configuration(self, nextcloud_handler, k8s_templates):

        sys.stdout.write("* Creating Nextcloud user {}\n".format(self.username))
        _created, _credentials = nextcloud_handler.create_user(
            self.email, self.username, self.first_name, self.last_name)

        if not _created:
            if KubernetesHandler.check_existence("secret", self.namespace, "nextcloud-credentials"):
                sys.stdout.write("* The user {} already exists in Nextcloud. Skipping\n".format(self.username))
                return

            sys.stdout.write("* The user {} already exists but the credentials are unknown. Updating password\n"
                             .format(self.username))
            _credentials = nextcloud_handler.update_password(self.username)

        sys.stdout.write("* Storing the credentials in Kubernetes\n")
        _username, _password = map(
            lambda str : base64.b64encode(str.encode("ascii")).decode("ascii"),
            _credentials)
        k8s_templates["nextcloudcredentials"].apply_template(
            namespace_name=self.namespace,
            username=_username, password=_password)

    def kubernetes_configuration(self, k8s_templates):
        # Create the Kubernetes resources
        sys.stdout.write("* Creating Kubernetes resources for user {}\n".format(self.username))

        # Namespace
        k8s_templates["namespace"].apply_template(
            namespace_name=self.namespace,
            rendered_name="{} {}".format(self.first_name, self.last_name).replace(" ", "_").replace("'", ""),
            namespace_type=Tenant.get_namespace_type())
        KubernetesHandler.add_resource_label(
            "namespaces", self.namespace, Course.get_group_name(self.course_code),
            "admin" if self.is_admin else "student")

        # Role binding
        k8s_templates["rolebinding-tenant"].apply_template(
            namespace_name=self.namespace,
            username=self.username)

        # Resource quota
        k8s_templates["resourcequota"].apply_template(
            namespace_name=self.namespace)

        ## Network Policies
        k8s_templates["ingress-netpol"].apply_template(
            namespace_name=self.namespace)


    def _create_user(self, keycloak_handler):
        if self.user is None:
            sys.stdout.write("* Creating user {} in Keycloak\n".format(self.username))
            self.user = keycloak_handler.create_new_user(
                self.email, self.username, self.first_name, self.last_name, self.namespace)
        else:
            sys.stdout.write("* User {} already exists in Keycloak\n".format(self.username))

        sys.stdout.write("* Checking user {} namespace attribute\n".format(self.username))
        keycloak_handler.add_namespace_attribute(self.user, self.namespace)

    def _assign_self_group(self, keycloak_handler):
        sys.stdout.write("* Creating group/role {} in Keycloak\n".format(self.group_name))
        keycloak_handler.create_group(self.group_name)
        keycloak_handler.create_client_role(self.group_name)

        sys.stdout.write("* Assigning user {} to group {}\n".format(self.username, self.group_name))
        keycloak_handler.add_user_to_group(self.user, self.group_name)

    def _assign_course_group(self, keycloak_handler):
        sys.stdout.write("* Assigning user {} to course {}\n".format(self.username, self.course_code))
        keycloak_handler.add_course_attribute(self.user, self.course_code)

        try:
            keycloak_handler.add_user_to_group(self.user, Course.get_group_name(self.course_code))
            if self.is_admin:
                keycloak_handler.add_user_to_group(self.user, Course.get_group_name_admin(self.course_code))
        except KeycloakHandler.GroupNotFound as _ex:
            sys.stderr.write("Impossible to assign group to {}: {}".format(self.username, _ex))
        except KeycloakHandler.ClientRoleNotFound as _ex:
            sys.stderr.write("Impossible to assign client role to {}: {}".format(self.username, _ex))

    @staticmethod
    def get_namespace_type():
        return "tenant"

    @staticmethod
    def get_namespace_name(username):
        return "tenant-{}".format(username.lower()).replace(".", "-")

    @staticmethod
    def get_group_name(username):
        return "tenant-{}".format(username.lower()).replace(".", "-")


class Laboratory:
    def __init__(self, record):
        self.course_code = record['Course (code)'].lower()
        self.number = record['Lab number']
        self.image = record['Image']
        self.cpu = record['Cpu']
        self.memory = record['Memory']
        self.description = record['Description']
        self.namespace = Course.get_namespace_name(self.course_code)
        self.vm_type = 'GUI' if record.get('Needs GUI', True) else 'CLI'

        sys.stdout.write("Processing laboratory {} - course {}\n"
                         .format(self.number, self.course_code))

    def kubernetes_configuration(self, k8s_templates):
        # Create the Kubernetes resources
        sys.stdout.write("* Creating Kubernetes resources for laboratory {} - course {}\n"
                         .format(self.number, self.course_code))

        k8s_templates["labtemplate"].apply_template(
            namespace_name=self.namespace, course_code=self.course_code, lab_number=self.number,
            description=self.description, image=self.image, vm_type=self.vm_type,
            cpu=self.cpu, memory=self.memory)


def _parse_csv_file(path):
    if path is None:
        return pd.DataFrame(columns=["Course (code)"])

    try:
        return pd.read_csv(path, comment='#')
    except FileNotFoundError:
        sys.stderr.write("Input CSV file ('{}') does not exist. Abort\n".format(path))
        sys.exit(1)
    except pd.errors.ParserError:
        sys.stderr.write("Impossible to parse CSV file ('{}'). Abort\n".format(path))
        sys.exit(1)

if __name__ == "__main__":
    # Parse the command line arguments
    _parser = argparse.ArgumentParser(description="Automatic creation of CrownLabs courses, laboratory and tenants")
    _parser.add_argument("keycloak_user", help="The admin username for the OIDC server")
    _parser.add_argument("keycloak_pass", help="The admin password for the OIDC server")
    _parser.add_argument("nextcloud_user", help="The admin username for Nextcloud server")
    _parser.add_argument("nextcloud_pass", help="The admin password for Nextcloud server")
    _parser.add_argument("-c", "--courses", metavar="<courses.csv>", help="The CSV file containing the courses to be created")
    _parser.add_argument("-l", "--laboratories", metavar="<laboratories.csv>", help="The CSV file containing the list of laboratories to be created")
    _parser.add_argument("-t", "--teachers", metavar="<teachers.csv>", help="The CSV file containing the professor accounts to be created")
    _parser.add_argument("-s", "--students", metavar="<students.csv>", help="The CSV file containing the student accounts to be created")

    _args = _parser.parse_args()

    # Parse the CSV files
    _courses = _parse_csv_file(_args.courses)
    _teachers = _parse_csv_file(_args.teachers)
    _students = _parse_csv_file(_args.students)
    _laboratories = _parse_csv_file(_args.laboratories)

    # Establish a connection with Keycloak
    try:
        sys.stdout.write("Establishing connection to Keycloak (user: {})\n".format(_args.keycloak_user))
        _keycloak_handler = KeycloakHandler(_args.keycloak_user, _args.keycloak_pass, "k8s")
        sys.stdout.write("Connection correctly established to Keycloak\n\n")
    except kc.exceptions.KeycloakAuthenticationError:
        sys.stderr.write("Invalid admin credentials (user: {}). Abort\n".format(_args.keycloak_user))
        sys.exit(1)
    except KeycloakHandler.ClientIDNotFound as _ex:
        sys.stderr.write("An error has occurred: {}. Abort\n".format(_ex))
        sys.exit(1)

    # Check kubectl availability and privileges
    try:
        KubernetesHandler.check_kubectl_cluster_admin()
    except KubernetesHandler.FatalError as _ex:
        sys.stderr.write("{}. Abort\n".format(_ex))
        sys.exit(1)


    # Prepare the templates for the resources to create in Kubernetes
    try:
        _k8s_templates = {
            "namespace": KubernetesTemplateHandler("namespace.yaml.tmpl", "templates/"),
            "clusterrolebinding-courseadmin": KubernetesTemplateHandler("clusterrolebinding-courseadmin.yaml.tmpl", "templates/"),
            "rolebinding-tenant": KubernetesTemplateHandler("rolebindingtenant.yaml.tmpl", "templates/"),
            "rolebinding-course": KubernetesTemplateHandler("rolebindingcourse.yaml.tmpl", "templates/"),
            "rolebinding-courseadmin": KubernetesTemplateHandler("rolebindingcourseadmin.yaml.tmpl", "templates/"),
            "resourcequota": KubernetesTemplateHandler("resourcequota.yaml.tmpl", "templates/"),
            "nextcloudcredentials": KubernetesTemplateHandler("nextcloudcredentials.yaml.tmpl", "templates/"),
            "labtemplate": KubernetesTemplateHandler("labtemplate.yaml.tmpl", "templates/"),
            "ingress-netpol": KubernetesTemplateHandler("ingress-netpol.yaml.tmpl", "templates/"),
        }
    except Template.exceptions.TemplateNotFound as _ex:
        sys.stderr.write("Failed to parse the Jinja2 template: '{}' does not exist. Abort\n".format(_ex))
        sys.exit(1)

    # Setup the Nextcloud communication handler
    _nextcloud_handler = NextcloudHandler(
        "https://crownlabs.polito.it/cloud",
        _args.nextcloud_user, _args.nextcloud_pass)

    # Iterate over the courses
    for _, _record in _courses.iterrows():
        _course = Course(_record)
        _course.keycloak_configuration(_keycloak_handler)
        _course.kubernetes_configuration(_k8s_templates)
        sys.stdout.write('\n')

    # Iterate over the teachers
    for _, _record in _teachers.iterrows():
        _teacher = Tenant(_record, is_admin=True)
        _teacher.keycloak_configuration(_keycloak_handler)
        _teacher.kubernetes_configuration(_k8s_templates)
        _teacher.nextcloud_configuration(_nextcloud_handler, _k8s_templates)

        sys.stdout.write('\n')

    # Iterate over the students
    for _, _record in _students.iterrows():
        _student = Tenant(_record)
        _student.keycloak_configuration(_keycloak_handler)
        _student.kubernetes_configuration(_k8s_templates)
        _student.nextcloud_configuration(_nextcloud_handler, _k8s_templates)
        sys.stdout.write('\n')

    # Iterate over the laboratories
    for _, _record in _laboratories.iterrows():
        _laboratory = Laboratory(_record)
        _laboratory.kubernetes_configuration(_k8s_templates)
        sys.stdout.write('\n')
