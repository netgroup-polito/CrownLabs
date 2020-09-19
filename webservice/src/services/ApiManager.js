import { Config, CustomObjectsApi, watch } from '@kubernetes/client-node';

/**
 * Class to manage all the interaction with the cluster
 *
 */
export default class ApiManager {
  /**
   * Constructor
   *
   * @param token the user token
   * @param type the token type
   * @param studentID the student id retrieved
   * @param templateNS the laboratories the user can see (cloud-computing, software-networking namespaces/roles in cluster)
   * @param instanceNS the user namespace where to run its instances
   */
  constructor(token, type, studentID, templateNS, instanceNS) {
    if (window.APISERVER_URL === undefined) {
      window.APISERVER_URL = APISERVER_URL;
    }

    this.kc = new Config(window.APISERVER_URL, token, type);
    this.apiCRD = this.kc.makeApiClient(CustomObjectsApi);
    this.templateGroup = 'crownlabs.polito.it';
    this.instanceGroup = 'crownlabs.polito.it';
    this.version = 'v1alpha1';
    this.templatePlural = 'labtemplates';
    this.instancePlural = 'labinstances';
    this.studentID = studentID;
    this.templateNamespace = templateNS;
    this.instanceNamespace = instanceNS;
  }

  async retrieveImageList() {
    return this.apiCRD
      .getClusterCustomObject(
        'crownlabs.polito.it',
        'v1alpha1',
        'imagelists',
        'crownlabs-virtual-machine-images'
      )
      .catch(error => {
        Promise.reject(error);
      });
  }

  /**
   * Private function called to retrieve all lab templates for a specific course (called by getCRDtemplates)
   *
   * @param course the specific course, the group the user belongs (cloud-computing, software-networking, ...)
   * @return the object {courseNamespace, List of lab templates} if available
   */
  async retrieveSingleCRDtemplate(course) {
    let ret = await this.apiCRD
      .listNamespacedCustomObject(
        this.templateGroup,
        this.version,
        course,
        this.templatePlural
      )
      .then(nodesResponse => {
        return nodesResponse.body.items.map(x => {
          return x.metadata.name;
        });
      })
      .catch(error => {
        Promise.reject(error);
      });
    if (ret !== null) {
      ret = { course, labs: ret };
    }
    return ret;
  }

  /**
   * Function to return all possible lab templates from all the group the user belongs (cloud-computing, software-networking, ...)
   * @returns {Promise<[unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown]>} the result of all the single
   * calls as a unique synchronized promise
   */
  async getCRDtemplates() {
    return Promise.all(
      this.templateNamespace.map(async x => this.retrieveSingleCRDtemplate(x))
    );
  }

  /**
   * Function to retrieve all lab instances in your namespace
   *
   * @returns the promise handling the request
   */
  getCRDinstances() {
    return this.apiCRD.listNamespacedCustomObject(
      this.instanceGroup,
      this.version,
      this.instanceNamespace,
      this.instancePlural
    );
  }

  /**
   * Function to create a lab instance
   *
   * @param labTemplateName the name of the lab template
   * @param labTemplateNamespace the namespace which the lab template belongs
   * @returns the promise handling the request
   */
  createCRDinstance(labTemplateName, labTemplateNamespace) {
    return this.apiCRD.createNamespacedCustomObject(
      this.instanceGroup,
      this.version,
      this.instanceNamespace,
      this.instancePlural,
      {
        apiVersion: `${this.instanceGroup}/${this.version}`,
        kind: 'LabInstance',
        metadata: {
          name: `${labTemplateName}-${this.studentID}-${
            Math.floor(Math.random() * 10000) + 1
          }`,
          namespace: this.instanceNamespace
        },
        spec: {
          labTemplateName,
          labTemplateNamespace,
          studentId: this.studentID
        }
      }
    );
  }

  /**
   * Function to delete a lab instance in your namespace
   *
   * @param name the name of the object you want to delete
   * @returns the promise handling the request
   */
  deleteCRDinstance(name, targetNamespace) {
    return this.apiCRD.deleteNamespacedCustomObject(
      this.instanceGroup,
      this.version,
      targetNamespace || this.instanceNamespace,
      this.instancePlural,
      name,
      {}
    );
  }

  deleteCRDtemplate(namespace, name) {
    return this.apiCRD.deleteNamespacedCustomObject(
      this.templateGroup,
      this.version,
      namespace,
      this.templatePlural,
      name,
      {}
    );
  }

  /**
   * Function to create a lab template (by a professor)
   * @param namespace the namespace where the template should be created
   * @param labNumber
   * @param description
   * @param cpu
   * @param memory
   * @param image
   */
  createCRDtemplate(namespace, labNumber, description, cpu, memory, image) {
    const courseName = namespace.split('course-')[1];

    return this.apiCRD.createNamespacedCustomObject(
      this.templateGroup,
      this.version,
      namespace,
      this.templatePlural,
      {
        apiVersion: `${this.templateGroup}/${this.version}`,
        kind: 'LabTemplate',
        metadata: {
          name: `${courseName}-lab${labNumber}`,
          namespace
        },
        spec: {
          courseName,
          description,
          labName: `${courseName}-lab${labNumber}`,
          labNum: labNumber,
          vm: {
            apiVersion: 'kubevirt.io/v1alpha3',
            kind: 'VirtualMachineInstance',
            metadata: {
              labels: {
                name: `${courseName}-lab${labNumber}`
              },
              name: `${courseName}-lab${labNumber}`,
              namespace
            },
            spec: {
              domain: {
                cpu: { cores: cpu },
                devices: {
                  disks: [
                    { disk: { bus: 'virtio' }, name: 'containerdisk' },
                    { disk: { bus: 'virtio' }, name: 'cloudinitdisk' }
                  ]
                },
                memory: { guest: `${memory}G` },
                resources: {
                  limits: { cpu: `${cpu + 1}`, memory: `${memory + 1}G` },
                  requests: { cpu: `${cpu}`, memory: `${memory}G` }
                }
              },
              volumes: [
                {
                  containerDisk: {
                    image,
                    imagePullSecret: 'registry-credentials'
                  },
                  name: 'containerdisk'
                },
                {
                  cloudInitNoCloud: {
                    secretRef: { name: `${courseName}-lab${labNumber}` }
                  },
                  name: 'cloudinitdisk'
                }
              ],
              terminationGracePeriodSeconds: 30
            }
          }
        }
      }
    );
  }

  /**
   * Function to watch events in the user namespace
   *
   * @param func the function to be called at each event
   * @param queryParam the query parameters you are going to use (Used when calling function to watch admin namespaces)
   */
  startWatching(func, queryParam = {}) {
    const path =
      Object.keys(queryParam).length !== 0
        ? `/apis/${this.instanceGroup}/${this.version}/${this.instancePlural}`
        : `/apis/${this.instanceGroup}/${this.version}/namespaces/${this.instanceNamespace}/${this.instancePlural}`;
    watch(
      this.kc,
      path,
      queryParam,
      (type, object) => {
        func(type, object);
      },
      e => {
        func(null, e);
      }
    );
  }

  /**
   * @@@@ UNUSED (since watcher has been patched and works)
   * Function to get a specific lab instance status
   *
   * @param name the name of the lab instance
   * @returns the promise handling the request
   */
  getCRDstatus(name) {
    return this.apiCRD.getNamespacedCustomObjectStatus(
      this.instanceGroup,
      this.version,
      this.instanceNamespace,
      this.instancePlural,
      name
    );
  }
}
