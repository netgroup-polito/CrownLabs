#!/usr/bin/env python3

import argparse
import logging
import sched
import time

import grequests
import kubernetes
import requests


class ImageListUpdater:
    """
    This class periodically requests of the list of images from the Docker registy
    and saves the obtained information as a Kubernetes object.
    """

    def __init__(self, image_list_requestor, image_list_saver, registry_adv_name):
        """
        Initializes the object.
        :param image_list_requestor: the handler used to request the list of images from the Docker registry.
        :param image_list_saver: the handler used to save the retrieved information as a Kubernetes object.
        :param registry_adv_name: the Docker registry host name advertised in the ImageList object.
        """

        self.image_list_requestor = image_list_requestor
        self.image_list_saver = image_list_saver
        self.registry_adv_name = registry_adv_name
        self.scheduler = sched.scheduler(time.time, time.sleep)

    def run_update_process(self, interval):
        """
        Starts the scheduler loop to request and save the image list.
        :param interval: The interval (in seconds) between one execution and the following.
        """

        self._run_periodically(interval, self.update)
        self.scheduler.run()

    def update(self):
        """
        Performs the actual update process.
        """

        start = time.time()
        logger.debug("Starting the update process")

        try:
            # Obtain the list of images from the Docker registry
            images = self.image_list_requestor.get_image_list()
        except (requests.exceptions.RequestException, ValueError):
            logger.exception("Failed to retrieve data from upstream")
            return

        try:
            self.image_list_saver.update_image_list({
                "registryName": self.registry_adv_name,
                "images": ImageListUpdater.__process_image_list(images),
            })
        except kubernetes.client.rest.ApiException:
            logger.exception("Failed to save data as ImageList")
            return

        logger.info(f"Update process correctly completed in {time.time() - start:.2f} seconds")

    def _run_periodically(self, interval, action, *args, **kwargs):
        """
        Runs a given action periodically.
        :param interval: the interval between multiple executions (in seconds).
        :param action: the function to be executed periodically.
        :param kwargs: keyworded arguments passed to the action function.
        """

        logger.debug(f"Executing '{action.__name__}': next scheduled in {interval} seconds")

        # Schedule the next execution
        self.scheduler.enter(
            delay=interval, priority=1, action=self._run_periodically,
            argument=(interval, action), kwargs=kwargs)

        # Execute the action
        action(**kwargs)

    @staticmethod
    def __process_image_list(images):
        """
        Processes the list of images returned from upstream to remove the "latest" tags
        and converts it to the correct format expected by Kubernetes.
        :param images: the list of images retrieved from the Docker registry.
        :return: a list of images, suitable to be saved as ImageList.
        """

        converted_images = []

        for image in images:

            # Get the available tags
            versions = image.get("tags") or []

            # Remove the "latest" tag
            try:
                versions.remove("latest")
            except ValueError:
                pass

            # Are there still any tags?
            if versions:
                converted_images.append({
                    "name": image.get("name"),
                    "versions": versions,
                })

        return converted_images


class ImageListRequestor:
    """
    This class interacts with the docker registry to get the list of images currently available.
    """

    def __init__(self, url, username=None, password=None):
        """
        Initializes the object.
        :param url: the url used to contact the Docker registry.
        :param username: the username used to access the Docker registry (optional).
        :param password: the password used to access the Docker registry (optional).
        """

        self.url = url
        self.auth = (username, password)

    def get_image_list(self):
        """
        Performs the requests upstream to retrieve the information about the images.
        :return: an object representing the information retrieved.
        """

        logger.debug("Requesting the registry catalog upstream")
        repositories = self._do_single_get(
            ImageListRequestor.__get_catalog_path()
        ).get("repositories", [])

        logger.debug("Requesting the image details upstream")
        images = self._do_parallel_gets(
            ImageListRequestor.__map_repositories_to_paths(repositories)
        )

        return images

    def _do_single_get(self, path):
        """
        Performs a GET to the target path and returns the result.
        :param path: the path to be retrieved.
        :return: the json object extracted from the response.
        """

        return requests.get(url=f"{self.url}{path}", auth=self.auth).json()

    def _do_parallel_gets(self, paths):
        """
        Performs a set of parallel GETs to the target paths and returns the results.
        :param paths: the paths to be retrieved.
        :return: the set of json object extracted from the response.
        """

        requests = (grequests.get(f"{self.url}{path}", auth=self.auth) for path in paths)
        return (response.json() for response in grequests.imap(requests))

    @staticmethod
    def __get_catalog_path():
        """
        Returns the URL path corresponding to the catalog.
        :return: the URL path corresponding to the catalog.
        """
        return "/v2/_catalog"

    @staticmethod
    def __map_repositories_to_paths(repositories):
        """
        Returns the URL paths to obtain detailed information about the repositories.
        :param repositories: the set of repositories of interest.
        :return: the URL paths corresponding to the input repositories.
        """
        return (f"/v2/{repo}/tags/list" for repo in repositories)


class ImageListSaver:
    """
    Saves the list of images retrieved from the Docker registry as a Kubernetes object.
    """

    def __init__(self, name):
        """
        Initializes the object and loads the Kubernetes configuration.
        :param name: The name assigned to the ImageList resource.
        """

        self.name = name

        try:
            # Configuration loaded from a kube config
            kubernetes.config.load_kube_config()
        except kubernetes.config.config_exception.ConfigException:
            # Configuration loaded from within a pod
            kubernetes.config.load_incluster_config()

    def update_image_list(self, image_list_spec):
        """
        Updates the content or creates a new ImageList object.
        :param image_list_spec: the content of the ImageList object to be updated.
        """

        resource_version = self._get_image_list_version()
        if resource_version:
            self._update_image_list(image_list_spec, resource_version)
        else:
            self._create_image_list(image_list_spec)

    def _get_image_list_version(self):
        """
        Gets the current version of the ImageList.
        :returns: the current version of the image list or None
        """

        api_instance = kubernetes.client.CustomObjectsApi()
        try:
            data = api_instance.get_cluster_custom_object(
                **ImageListSaver.__get_imagelist_args(), name=self.name
            )
        except kubernetes.client.rest.ApiException:
            return None

        resource_version = data.get("metadata", {}).get("resourceVersion", None)
        logger.debug(f"Retrieved ImageList resource version: '{resource_version}'")
        return resource_version

    def _create_image_list(self, image_list_spec):
        """
        Creates a new ImageList object.
        :param image_list_spec: the content of the ImageList object to be created.
        """

        api_instance = kubernetes.client.CustomObjectsApi()
        api_instance.create_cluster_custom_object(
            **ImageListSaver.__get_imagelist_args(),
            body=self._create_image_list_object(image_list_spec)
        )
        logger.debug(f"ImageList '{self.name}' correctly created'")

    def _update_image_list(self, image_list_spec, resource_version):
        """
        Updates an existing ImageList object.
        :param image_list_spec: the content of the ImageList object to be updated.
        :param resource_version: the version of the resource to be updated.
        """

        api_instance = kubernetes.client.CustomObjectsApi()
        api_instance.replace_cluster_custom_object(
            **ImageListSaver.__get_imagelist_args(), name=self.name,
            body=self._create_image_list_object(image_list_spec, resource_version)
        )
        logger.debug(f"ImageList '{self.name}' correctly updated'")

    def _create_image_list_object(self, image_list_spec, resource_version=None):
        """
        Creates a new ImageList object, given the spec body.
        :param image_list_spec: the content of the ImageList object to be created.
        :param resource_version: the version of the resource to be updated.
        :returns: the ImageList json representation.
        """

        return {
            "apiVersion": "crownlabs.polito.it/v1alpha1",
            "kind": "ImageList",
            "metadata": {
                "name": self.name,
                "resourceVersion": resource_version,
            },
            "spec": image_list_spec,
        }

    @staticmethod
    def __get_imagelist_args():
        """
        Returns the parameters describing the ImageList API.
        """

        return {"group": "crownlabs.polito.it", "version": "v1alpha1", "plural": "imagelists"}


# Initialize the logger object
logger = logging.getLogger("update-crownlabs-image-list")


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
    parser = argparse.ArgumentParser(
        description="Periodically requests the list of images from a Docker registry and stores it as a Kubernetes CR")

    parser.add_argument("--advertised-registry-name", required=True,
                        help="the host name of the Docker registry where the images can be retrieved")
    parser.add_argument("--image-list-name", required=True,
                        help="the name assigned to the resulting ImageList object")
    parser.add_argument("--registry-url", required=True, help="the URL used to contact the Docker registry")
    parser.add_argument("--registry-username", help="the username used to access the Docker registry")
    parser.add_argument("--registry-password", help="the password used to access the Docker registry")
    parser.add_argument("--update-interval", required=True, type=int,
                        help="the interval (in seconds) between one update and the following")

    args = parser.parse_args()

    # Create the object reading the list of the images from the Docker registry
    logger.info(f"Upstream Docker registry: '{args.registry_url}' - Username: '{args.registry_username}'")
    image_list_requestor = ImageListRequestor(args.registry_url, args.registry_username, args.registry_password)

    # Create the object saving the retrieved information as a Kubernetes object
    logger.info(f"Target ImageList object: '{args.image_list_name}'")
    image_list_saver = ImageListSaver(args.image_list_name)

    # Create the object periodically perforing the update process
    image_list_updater = ImageListUpdater(image_list_requestor, image_list_saver, args.advertised_registry_name)

    logger.info("Starting the update process")
    try:
        image_list_updater.run_update_process(args.update_interval)
    except KeyboardInterrupt:
        logger.info("Received stop signal. Exiting")
