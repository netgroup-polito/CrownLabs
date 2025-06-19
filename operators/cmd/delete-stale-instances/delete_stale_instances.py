from kubernetes import client, config
from kubernetes.client.rest import ApiException
from datetime import datetime
import argparse
import math
import logging
import re


class InstanceExpiredDeleter:
    """
    This class interacts with crownlabs instances and templates to delete expired instances.
    """
    group = "crownlabs.polito.it"
    version = "v1alpha2"
    plural_instance = "instances"
    plural_template = "templates"

    delete_after_regex = re.compile("([0-9]+)([mhd])")

    def __init__(self, dry_run=None):
        """
        Initializes the object.
        :param dry_run: application run mode. Default value False.
        """

        self.dry_run = dry_run
        try:
            # Configuration loaded from a kube config
            config.load_kube_config()
        except config.config_exception.ConfigException:
            # Configuration loaded from within a pod
            config.load_incluster_config()

    def get_instance_list(self):
        """
        Gets the current list of instances.
        :returns: the list of instances or None
        """
        # connect to Custom Object Api
        api_instance = client.CustomObjectsApi()
        try:
            instances = api_instance.list_cluster_custom_object(group=self.group, version=self.version,
                                                                plural=self.plural_instance, pretty='pretty')
            logger.debug(f"Instances list retrieved successfully ({len(instances.get('items'))} items)")
        except ApiException as e:
            logger.error(f"Failed retrieving the list of instances: '{e}'")
            return None
        return instances

    @staticmethod
    def instance_is_expired(lifespan, timecreation):
        """
        Compares creation timestamp with current timestamp to decide if instance is expired.
        :returns: True if expired or False if it is not
        :lifespan: instance life span in seconds
        :timecreation: instance creation timestamp
        """
        now = datetime.now()
        # calculate time difference to verify expiration status
        deltatime = now - datetime.strptime(timecreation, '%Y-%m-%dT%H:%M:%SZ')
        total_time = deltatime.total_seconds()
        return (total_time > lifespan)

    @staticmethod
    def convert_to_life_span(delete_after):
        """
        Converts the delete_after string to the corresponding value in seconds.
        :returns: delete_after paramenter converted to seconds.
        :delete_after: string coming from template rappresenting the expiration threshold of the instance.
        :delete_after: has a standard format of [0-9]+[mhd]
        """

        # Do not terminate the instances with delete_after policy "never"
        if delete_after == "never":
            return math.inf

        delete_after_match = InstanceExpiredDeleter.delete_after_regex.match(delete_after)
        if delete_after_match is None:
            logger.error(f"DeleteAfter field has a wrong format: '{delete_after}'")
            return None
        delete_after_vector = delete_after_match.groups()
        multiplier = 0
        if delete_after_vector[1] == 'm':
            multiplier = 60
        elif delete_after_vector[1] == 'h':
            multiplier = 60 * 60
        elif delete_after_vector[1] == 'd':
            multiplier = 60 * 60 * 24
        else:
            logger.error(f"DeleteAfter field has a wrong time unit: '{delete_after_vector[1]}'")
            return None
        time = int(delete_after_vector[0]) * multiplier
        logger.debug(f"Time converted successfully: {delete_after} = {time} seconds")
        return time

    def get_life_span(self, template_name, template_ns):
        """
        Retrieves deleteAfter field of template specified.
        :returns: life span in seconds or None
        :template_name: template name
        :template_ns: template namespace
        """
        api_instance = client.CustomObjectsApi()
        try:
            template = api_instance.get_namespaced_custom_object(group=self.group, version=self.version,
                                                                 namespace=template_ns, plural=self.plural_template,
                                                                 name=template_name)
        except ApiException as e:
            logger.error(f"Failed retrieving template {template_ns}/{template_name}: '{e}'")
            return None

        delete_after = template.get("spec").get("deleteAfter")
        lifespan = InstanceExpiredDeleter.convert_to_life_span(delete_after)
        logger.debug(f"Retrieved template: '{template_name}' in namespace: '{template_ns}' "
                     f"with maximum lifetime: '{delete_after}' seconds")
        return lifespan

    def delete_instance_expired(self, instances):
        """
        Deletes from current list of instances the expired ones.
        :returns: None
        :instances: current instance list in crowlabs
        """
        api_instance = client.CustomObjectsApi()
        for instance in instances.get("items"):
            # get instance name
            name = instance.get("metadata").get("name")
            # get instance creation timestamp
            creation_timestamp = instance.get("metadata").get("creationTimestamp")
            # get instance
            namespace = instance.get("metadata").get("namespace")
            # retrieve template name
            template_name = instance.get("spec").get("template.crownlabs.polito.it/TemplateRef").get("name")
            # retrieve template namespace
            template_ns = instance.get("spec").get("template.crownlabs.polito.it/TemplateRef").get("namespace", "default")

            # retrieve instance life span
            logger.debug(f"Processing instance: '{name}' in namespace: '{namespace}' created at: '{creation_timestamp}'")
            lifespan = self.get_life_span(template_name, template_ns)
            if lifespan is None:
                logger.error(f"Template: '{template_name}' in namespace: '{template_ns}' "
                             "has a wrong delete_after field format")
            else:
                # verify instance expiration status
                if InstanceExpiredDeleter.instance_is_expired(lifespan, creation_timestamp):
                    try:
                        # delete expired instance
                        api_instance.delete_namespaced_custom_object(group=self.group, version=self.version,
                                                                     namespace=namespace, plural=self.plural_instance,
                                                                     name=name, dry_run=self.dry_run)
                        dry_run_str = "(dry-run)" if self.dry_run else ""
                        logger.debug(f"Deleted instance: '{name}' in namespace: '{namespace}' {dry_run_str}")
                    except ApiException as e:
                        # exception occurred while deleting instance
                        logger.error(f"Failed to delete instance {namespace}/{name}: '{e}'")
                else:
                    logger.debug(f"Instance: '{name}' in namespace: '{namespace}' not yet expired")


# Initialize the logger object
logger = logging.getLogger("Delete stale CrownLabs instances")


def configure_logger():
    """
    Configures the logger object with the required configuration
    """

    logger.setLevel(logging.DEBUG)

    formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")

    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.DEBUG)
    console_handler.setFormatter(formatter)

    logger.handlers.clear()
    logger.addHandler(console_handler)


if __name__ == "__main__":

    # Configure the logger
    configure_logger()

    # Parse the command line arguments
    parser = argparse.ArgumentParser(description="Periodically check instances expiration")

    parser.add_argument("--dry-run", required=False, nargs='?', const='All', default=None,
                        help="option dry-run")
    args = parser.parse_args()

    logger.info(f"Deletion mode dry-run is: '{args.dry_run}'")
    instance_expired_deleter = InstanceExpiredDeleter(args.dry_run)
    logger.info("Starting the deletion process")
    try:
        instances = instance_expired_deleter.get_instance_list()
        instance_expired_deleter.delete_instance_expired(instances)
        logger.info("Deletion process completed correctly")
    except KeyboardInterrupt:
        logger.info("Received stop signal. Exiting")
