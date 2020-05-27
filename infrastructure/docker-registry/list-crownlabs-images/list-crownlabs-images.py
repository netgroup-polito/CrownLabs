import argparse
import json
import logging
import time

from functools import partial
from http.server import ThreadingHTTPServer, BaseHTTPRequestHandler
from threading import Lock

import grequests
import requests


# Initialize the logger object
logger = logging.getLogger("list-crownlabs-images")

class RequestWrapper:
    """
    This class represents a wrapper used to actually interact with the docker registry.
    """

    def __init__(self, url, username=None, password=None):
        """
        Initializes the wrapper object.
        :param url: the url used to contact the Docker registry.
        :param username: the username used to access the Docker registry.
        :param password: the password used to access the Docker registry.
        """

        self.url = url
        self.auth = (username, password)

    def do_get(self, path):
        """
        Performs a GET to the target path and returns the result.
        :param path: the path to be retrieved.
        :return: the json object extracted from the response.
        """

        return requests.get(url=f"{self.url}{path}", auth=self.auth).json()

    def do_parallel_gets(self, paths):
        """
        Performs a set of parallel GETs to the target paths and returns the results.
        :param paths: the paths to be retrieved.
        :return: the set of json object extracted from the response.
        """

        requests = (grequests.get(f"{self.url}{path}", auth=self.auth) for path in paths)
        return (response.json() for response in grequests.imap(requests))


class ImagesListHandler:
    """
    This class represents the handler requesting and caching the information retrieved from the Docker registry.
    """

    def __init__(self, request_wrapper, cache_duration, adv_registry_name):
        """
        Initializes the handler object.
        :param request_wrapper: the object used ot perform the actual requests.
        :param cache_duration: the amount of time (in minutes) the cached data is valid.
        :param adv_registry_name: tha name of the registry advertised to the clients.
        """

        self.adv_registry_name = adv_registry_name
        self.cache_duration = cache_duration * 60
        self.request_wrapper = request_wrapper

        self.cache = None
        self.cached_time = None
        self.lock = Lock()

    def get(self):
        """
        Returns the information retrieved from the upstream (either obtained or cached)
        :return: an object representing the information retrieved.
        """

        with self.lock:
            if self._is_cache_expired():
                logger.info("Cache is invalid or expired: requesting upstream")
                self._update_cache(self._do_real_get())
            else:
                logger.info("Returning cached data")

            return self.cache

    def _do_real_get(self):
        """
        Performs the requests to the upstream to retrieve the information about the images
        :return: an object representing the information retrieved.
        """

        logger.debug("Requesting registry catalog to upstream")
        repositories = self.request_wrapper.do_get(
            ImagesListHandler.__get_catalog_path()
        ).get("repositories", [])

        logger.debug("Requesting image details to upstream")
        images = self.request_wrapper.do_parallel_gets(
            ImagesListHandler.__map_repositories_to_paths(repositories)
        )

        return {
            "registry_name": self.adv_registry_name,
            "images": self._filter_images_latest(images),
        }

    def _filter_images_latest(self, images):
        """
        Removes the "latest" tag from the lists and discards images without tags;
        :param images: the list of images to be analysed;
        :return: an ordered list of images, without the "latest" tags.
        """

        filtered_images = []

        for image in sorted(images, key=lambda image: image.get("name")):

            # Remove the "latest" tags
            try:
                image.get("tags", []).remove("latest")
            except ValueError:
                pass

            # Are there still any tags?
            if len(image.get("tags", [])):
                filtered_images.append(image)

        return filtered_images

    def _is_cache_expired(self):
        """
        Returns whether the cached data is expired.
        :return: whether the cached data is expired.
        """

        return self.cache is None or self.cached_time is None \
            or time.time() - self.cached_time > self.cache_duration

    def _update_cache(self, data):
        """
        Updates the data stored in cache
        :param data: the data to be cached.
        """

        self.cache = data
        self.cached_time = time.time()

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


class HTTPRequestHandler(BaseHTTPRequestHandler):
    """
    This class defines how to handle the requests from the clients.
    """

    def __init__(self, images_list_handler, *args, **kwargs):
        """
        Initializes the handler with the object iteracting with the docker registry.
        :param images_list_handler: the handler used to interact with the docker registry.
        """

        self.images_list_handler = images_list_handler

        # BaseHTTPRequestHandler calls do_GET **inside** __init__ !!!
        # So we have to call super().__init__ after setting attributes.
        super().__init__(*args, **kwargs)

    def do_GET(self):
        """
        Handles the GET requests from the clients
        """

        if self.path == "/healthz":
            self._do_GET_healthz()
            return

        start = time.time()
        logger.debug("Start handling GET request")

        try:
            data = self.images_list_handler.get()
        except (requests.exceptions.RequestException, ValueError):
            logger.exception("Failed to retrieve data from upstream")
            self.send_response(502)
            self.end_headers()
            return

        # Set the response headers
        self.send_response(200)
        self.send_header('Content-type', 'application/json; charset=utf-8')
        self.end_headers()

        self.wfile.write(json.dumps(data).encode(encoding='utf_8'))

        logger.debug(f"GET request handled in {time.time() - start:.2f} seconds")
        return

    def _do_GET_healthz(self):
        logger.debug(f"Answering to readiness probe")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"healthy\n")


def configure_logger():
    """
    Configures the logger object with the required configuration
    """

    logger.setLevel(logging.DEBUG)

    formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")

    console_handler = logging.StreamHandler()
    console_handler.setLevel(logging.DEBUG)
    console_handler.setFormatter(formatter)

    logger.addHandler(console_handler)


if __name__ == "__main__":

    # Configure the logger
    configure_logger()

    # Parse the command line arguments
    parser = argparse.ArgumentParser(description="A simple web-server returning the list of images available on an upstream Docker registry")
    parser.add_argument("--webserver-port", help="the port the web-server is listening to", required=True, type=int)
    parser.add_argument("--registry-url", help="the URL used to contact the Docker registry", required=True)
    parser.add_argument("--registry-username", help="the username used to access the Docker registry")
    parser.add_argument("--registry-password", help="the password used to access the Docker registry")
    parser.add_argument("--advertised-registry-name", help="the name of the registry returned together with the images")
    parser.add_argument("--cache-duration-minutes", help="the amount of time the responces are cached (in minutes)", default=0, type=int)
    args = parser.parse_args()

    # Create the handlers to manage the requests
    request_wrapper = RequestWrapper(args.registry_url, args.registry_username, args.registry_password)
    images_list_handler = ImagesListHandler(
        request_wrapper, args.cache_duration_minutes,
        args.advertised_registry_name or args.registry_url)
    handler = partial(HTTPRequestHandler, images_list_handler)

    # Create the web-server
    server = ThreadingHTTPServer(("", args.webserver_port), handler)
    logger.info(f"Server listening on :{args.webserver_port}")

    # Start the web-server
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Received stop signal. Exiting")
