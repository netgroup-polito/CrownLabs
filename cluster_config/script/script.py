import json
import os
import argparse
import pandas as pd
import jinja2 as Template
from keycloak import KeycloakAdmin, exceptions

class KeycloakHandler:
    def __init__(self, admin_user, admin_pass, client_id):
        self.keycloak_admin = KeycloakAdmin(
            server_url="https://auth.crown-labs.ipv6.polito.it/auth/",
            username=admin_user,
            password=admin_pass,
            user_realm_name="master",
            realm_name="crownlabs",
            verify=True)
        self.client_id = self.keycloak_admin.get_client_id(client_id)

    def get_user(self, username):
        return self.keycloak_admin.get_users({'username': username})

    def create_new_user(self, email, username, first_name, last_name, course_code):
        _user = self.keycloak_admin.create_user({
            'email': email,
            'username': username,
            'firstName': first_name,
            'lastName': last_name,
            'enabled': True,
            'emailVerified': True,
            'attributes': {
                'namespace': "student-{}".format(username),
                'courses': [course_code, ],
            }
        })
        self.keycloak_admin.send_update_account(user_id=_user, payload=json.dumps(['UPDATE_PASSWORD']))

    def add_user_to_course(self, user_json, username, course_code):
        _user_id = user_json.get('id')
        _course_codes = list(set(user_json.get('attributes').get('courses') + [course_code, ]))
        return self.keycloak_admin.update_user(user_id=_user_id, payload={'attributes': {'namespace': "student-{}".format(username), 'courses' : _course_codes}})

    def add_client_role(self, course_code):
        try:
            self.keycloak_admin.get_client_role(self.client_id, course_code)
        except exceptions.KeycloakGetError:
            self.keycloak_admin.create_client_role(client_role_id=self.client_id, payload={'name': course_code, 'clientRole': True})


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
        os.system("echo '{}' | kubectl create -f -".format(_namespace))
        os.system("echo '{}' | kubectl create -f -".format(_privileges))
        os.system("echo '{}' | kubectl create -f -".format(_secret))


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

    _keycloak_handler = KeycloakHandler(_args.keycloak_user, _args.keycloak_pass, "k8s")
    _kubernetes_handler = KubernetesHandler(_template_namespace_path, _template_rolebinding_path, _template_regcred_path)

    parsed_csv_file = pd.read_csv(_args.csv_file, sep=",")
    for _, student in parsed_csv_file.iterrows():

        _username = student['Username']
        _email = student['Email']
        _last_name = student['Last name']
        _first_name = student['First name']
        _course_code = student['Course (code)']

        _keycloak_handler.add_client_role(_course_code)
        _user = _keycloak_handler.get_user(_username)

        if len(_user) == 0:
            # Create a new user
            print("\nProcessing user {} ({} {}) enrolled in {}".format(_username, _first_name, _last_name, _course_code))
            _keycloak_handler.create_new_user(_email, _username, _first_name, _last_name, _course_code)
            # Create the Kubernetes resources
            _kubernetes_handler.create_resources(_username, _course_code)

        else:
            # Update course_code
            print("\nUpdating user {} ({} {}) enrolled in {}".format(_username, _first_name, _last_name, _course_code))
            _keycloak_handler.add_user_to_course(_user[0], _username, _course_code)