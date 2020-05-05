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
    this.templateGroup = 'template.crown.team.com';
    this.instanceGroup = 'instance.crown.team.com';
    this.version = 'v1';
    this.templatePlural = 'labtemplates';
    this.instancePlural = 'labinstances';
    this.studentID = studentID;
    this.templateNamespace = templateNS;
    this.instanceNamespace = instanceNS;
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
      ret = { course: course, labs: ret };
    }
    return ret;
  }

  /**
   * Function to return all possible lab templates from all the group the user belongs (cloud-computing, software-networking, ...)
   * @returns {Promise<[unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown, unknown]>} the result of all the single
   * calls as a unique synchronized promise
   */
  async getCRDtemplates() {
    return await Promise.all(
      this.templateNamespace.map(
        async x => await this.retrieveSingleCRDtemplate(x)
      )
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
        apiVersion: this.instanceGroup + '/' + this.version,
        kind: 'LabInstance',
        metadata: {
          name:
            labTemplateName +
            '-' +
            this.studentID +
            '-' +
            (Math.floor(Math.random() * 10000) + 1),
          namespace: this.instanceNamespace
        },
        spec: {
          labTemplateName: labTemplateName,
          labTemplateNamespace: labTemplateNamespace,
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
  deleteCRDinstance(name) {
    return this.apiCRD.deleteNamespacedCustomObject(
      this.instanceGroup,
      this.version,
      this.instanceNamespace,
      this.instancePlural,
      name,
      {}
    );
  }

  /**
   * Function to create a lab template (by a professor)
   * @param course_code
   * @param lab_number
   * @param description
   * @param cpu
   * @param memory
   * @param image
   * @param namespace the namespace where the template should be created
   */
  createCRDtemplate(namespace, lab_number, description, cpu, memory, image) {
    return this.apiCRD.createNamespacedCustomObject(
      this.templateGroup,
      this.version,
      namespace,
      this.templatePlural,
      {
        apiVersion: this.instanceGroup + '/' + this.version,
        kind: 'LabTemplate',
        metadata: {
          name: namespace + '-lab' + lab_number,
          namespace: namespace
        },
        spec: {
          courseName: namespace,
          description: description,
          labName: namespace + '-lab' + lab_number,
          vm: {
            apiVersion: 'kubevirt.io/v1alpha3',
            kind: 'VirtualMachineInstance',
            metadata: {
              name: namespace + '-lab' + lab_number,
              namespace: namespace
            },
            labels: { name: namespace + '-lab' + lab_number },
            spec: {
              domain: {
                cpu: { cores: cpu },
                devices: {
                  disks: {
                    disk1: { bus: 'virtio', name: 'containerdisk' },
                    disk2: { bus: 'virtio', name: 'cloudinitdisk' }
                  },
                  memory: { guest: { memory } },
                  resource: {
                    limits: { cpu: cpu + 1, memory: memory + 0.5 },
                    request: { cpu: cpu + 0.5, memory: memory + 'G' }
                  }
                },
                terminationGracePeriodSeconds: 30,
                volumes: {
                  ContainerDisk: {
                    image: image,
                    imagePullSecret: 'registry-credentials'
                  },
                  cloudInitNoCloud: {
                    secretRef: { name: namespace + '-lab' + lab_number },
                    name: 'cloudinitdisk'
                  }
                }
              }
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
    let path =
      Object.keys(queryParam).length !== 0
        ? '/apis/' +
          this.instanceGroup +
          '/' +
          this.version +
          '/' +
          this.instancePlural
        : '/apis/' +
          this.instanceGroup +
          '/' +
          this.version +
          '/namespaces/' +
          this.instanceNamespace +
          '/' +
          this.instancePlural;
    watch(
      this.kc,
      path,
      queryParam,
      function (type, object) {
        func(type, object);
      },
      function (e) {
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
