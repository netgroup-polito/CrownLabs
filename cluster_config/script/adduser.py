import argparse
import os
import sys
import jinja2 as Template
import keycloak as kc
import pandas as pd


class KeycloakHandlerException(Exception):
    pass


class KeycloakHandler:
    def __init__(self, admin_user, admin_pass, client_id):
        self.keycloak_admin = kc.KeycloakAdmin(
            server_url="https://auth.crown-labs.ipv6.polito.it/auth/",
            username=admin_user,
            password=admin_pass,
            user_realm_name="master",
            realm_name="crownlabs",
            verify=True)

        self.client_id = self.keycloak_admin.get_client_id(client_id)
        if not self.client_id:
            raise KeycloakHandlerException("Client ID '{}' not found in Keycloak".format(client_id))

        self.client_roles = dict()

    def get_user(self, username):
        _users = self.keycloak_admin.get_users({'username': username})
        return None if len(_users) == 0 else _users[0]

    def create_new_user(self, email, username, first_name, last_name):
        _user_data = {
            'email': email,
            'username': username,
            'firstName': first_name,
            'lastName': last_name,
            'enabled': True,
            'emailVerified': True,
            'attributes': {
                'namespace': "student-{}".format(username),
            }
        }

        _user_id = self.keycloak_admin.create_user(_user_data)
        self.keycloak_admin.send_update_account(user_id=_user_id, payload='["UPDATE_PASSWORD"]')

        _user_data['id'] = _user_id
        return _user_data

    def add_user_to_course(self, user, course_code):
        _user_id = user.get('id')
        _attributes = user.get('attributes', {})
        _courses = _attributes.get('courses', [])
        _attributes['courses'] = list(set(_courses + [course_code, ]))

        self.keycloak_admin.update_user(user_id=_user_id, payload={'attributes': _attributes})
        self.keycloak_admin.assign_client_role(user_id=_user_id, client_id=self.client_id, roles=self.client_roles[course_code])

    def add_client_role(self, course_code):
        # Do not try to add the client role if already stored locally
        if course_code in self.client_roles.keys():
            return

        try:
            # Retrieve the client role
            self.client_roles[course_code] = self.keycloak_admin.get_client_role(self.client_id, course_code)
        except kc.exceptions.KeycloakGetError:
            # Create a new client role
            self.keycloak_admin.create_client_role(
                client_role_id=self.client_id,
                payload={'name': course_code, 'clientRole': True})
            # Retrieve the newly created client role
            self.client_roles[course_code] = self.keycloak_admin.get_client_role(self.client_id, course_code)


class KubernetesHandler:
    def __init__(self, template_namespace_path, template_rolebinging_file, template_regcred_file):
        _template_loader = Template.FileSystemLoader(searchpath="./")
        _template_env = Template.Environment(loader=_template_loader)
        self.template_ns = _template_env.get_template(template_namespace_path)
        self.template_rb = _template_env.get_template(template_rolebinging_file)
        self.template_sc = _template_env.get_template(template_regcred_file)

    def create_resources(self, username, course_code):
        _namespace = self.template_ns.render(namespace_name="student-{}".format(username))
        _privileges = self.template_rb.render(
            namespace_name="student-{}".format(username),
            course_code="kubernetes:{}".format(course_code), username=username)

        _secret = self.template_sc.render(namespace_name="student-{}".format(username))
        os.system("echo '{}' | kubectl apply -f - | sed 's/^/*** k8s: /g'".format(_namespace))
        os.system("echo '{}' | kubectl apply -f - | sed 's/^/*** k8s: /g'".format(_privileges))
        os.system("echo '{}' | kubectl apply -f - | sed 's/^/*** k8s: /g'".format(_secret))


if __name__ == "__main__":
    # Parse the command line arguments
    _parser = argparse.ArgumentParser()
    _parser.add_argument("csv_file", help="The CSV file containing the students to be inserted in Keycloak")
    _parser.add_argument("keycloak_user", help="The admin username for keycloak")
    _parser.add_argument("keycloak_pass", help="The admin password for keycloak")
    _args = _parser.parse_args()

    _template_namespace_path = 'namespace_template.yaml'
    _template_rolebinding_path = 'role_binding_template.yaml'
    _template_regcred_path = 'regcred_template.yaml'

    # Parse the input CSV file
    try:
        _parsed_csv_file = pd.read_csv(_args.csv_file)
    except FileNotFoundError:
        sys.stderr.write("Input CSV file ('{}') does not exist. Abort\n".format(_args.csv_file))
        sys.exit(1)
    except pd.errors.ParserError:
        sys.stderr.write("Impossible to parse CSV file ('{}'). Abort\n".format(_args.csv_file))
        sys.exit(1)

    # Establish a connection with Keycloak
    try:
        _keycloak_handler = KeycloakHandler(_args.keycloak_user, _args.keycloak_pass, "k8s")
    except kc.exceptions.KeycloakAuthenticationError:
        sys.stderr.write("Invalid admin credentials (user: {}). Abort\n".format(_args.keycloak_user))
        sys.exit(1)
    except KeycloakHandlerException as _ex:
        sys.stderr.write("An error has occurred: {}. Abort\n".format(_ex))
        sys.exit(1)

    # Prepare the templates for the resources to create in Kubernetes
    try:
        _kubernetes_handler = KubernetesHandler(_template_namespace_path, _template_rolebinding_path, _template_regcred_path)
    except Template.exceptions.TemplateNotFound as _ex:
        sys.stderr.write("Failed to parse the Jinja2 template: '{}' does not exist. Abort\n".format(_ex))
        sys.exit(1)

    for _, student in _parsed_csv_file.iterrows():

        _username = student['Username']
        _email = student['Email']
        _last_name = student['Last name']
        _first_name = student['First name']
        _course_code = student['Course (code)']

        sys.stdout.write("Processing user {} ({} {}) - Course: {}\n".format(_username, _first_name, _last_name, _course_code))

        # Grab the information about the user from Keycloak
        _user = _keycloak_handler.get_user(_username)

        # The user does not yet exist
        if _user is None:
            # Create a new user in Keycloak (automatically skipped if the user already exists)
            sys.stdout.write("* Creating user {} in Keycloak\n".format(_username))
            _user = _keycloak_handler.create_new_user(_email, _username, _first_name, _last_name)
        else:
            sys.stdout.write("* User {} already exists in Keycloak\n".format(_username))

        sys.stdout.write("* Ensuring course {} exists in Keycloak\n".format(_course_code))
        _keycloak_handler.add_client_role(_course_code)

        sys.stdout.write("* Assigning user {} to course {}\n".format(_username, _course_code))
        _keycloak_handler.add_user_to_course(_user, _course_code)

        # Create the Kubernetes resources
        sys.stdout.write("* Creating Kubernetes resources for user {}\n".format(_username))
        _kubernetes_handler.create_resources(_username, _course_code)

        sys.stdout.write('\n')
