import React from 'react';
import Toastr from 'toastr';
import Footer from './components/Footer';
import Header from './components/Header';
import ApiManager from './services/ApiManager';
import 'toastr/build/toastr.min.css';
import Body from './components/Body';

/**
 * Function to parse a JWT token
 * @param token the token received by keycloak
 * @returns {any} the decrypted token as a JSON object
 */
function parseJWTtoken(token) {
  const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
  return JSON.parse(
    decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => {
          return `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`;
        })
        .join('')
    )
  );
}

/**
 * Function to check the token, but encoded and decoded
 * @param parsed the decoded one
 * @return {boolean} true or false whether the token satisfies the constraints
 */
function checkToken(parsed) {
  if (!parsed.groups || !parsed.groups.length) {
    Toastr.error('You do not belong to any namespace to see laboratories');
    return false;
  }
  if (!parsed.namespace || !parsed.namespace[0]) {
    Toastr.error(
      'You do not have your own namespace where to run laboratories'
    );
    return false;
  }
  return true;
}

/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserLogic extends React.Component {
  /* The State variable contains:
   * - all lab templates as a Map: (course_group => Array of available templates for that course)
   * - all lab instances as a Map: (instance_name => URL if running, null otherwise)
   * - all ADMIN lab templates as a Map: (course_group => Array of available templates for that course)
   * - all ADMIN lab instances as a Map: (instance_name => URL if running, null otherwise)
   * - current selected CRD template as an object (name, namespace).
   * - current selected CRD instance
   * - all namespaced events as a string
   * - all ADMIN namespaced events as a string
   * - boolean variable whether to show the status info area
   * - adminHidden whether to render or not the admin page (changed by the button in the StudentView IF adminGroups is not false)
   * */

  constructor(props) {
    super(props);
    const { idToken, tokenType } = props;
    this.connect = this.connect.bind(this);
    this.retriveImageList = this.retriveImageList.bind(this);
    this.changeSelectedCRDtemplate = this.changeSelectedCRDtemplate.bind(this);
    this.changeSelectedCRDinstance = this.changeSelectedCRDinstance.bind(this);
    this.startCRDinstance = this.startCRDinstance.bind(this);
    this.stopCRDinstance = this.stopCRDinstance.bind(this);
    this.deleteCRDtemplate = this.deleteCRDtemplate.bind(this);
    this.createCRDtemplate = this.createCRDtemplate.bind(this);
    this.stopCRDinstanceAdmin = this.stopCRDinstanceAdmin.bind(this);
    this.notifyEvent = this.notifyEvent.bind(this);
    this.connectAdmin = this.connectAdmin.bind(this);
    this.notifyEventAdmin = this.notifyEventAdmin.bind(this);
    const parsedToken = parseJWTtoken(idToken);
    this.theme = false;
    if (!checkToken(parsedToken)) {
      this.logoutInterval();
    }
    /* Differentiate the two different kind of group: where the user is admin (professor or PhD) and the one where he is just a student */
    const adminGroups = parsedToken.groups
      .filter(x => x.match(/kubernetes:\S+admin/g))
      .map(x => x.replace('kubernetes:', '').replace('-admin', ''));
    const userGroups = parsedToken.groups
      .filter(x => x.includes('kubernetes:') && !x.includes('-admin'))
      .map(x => x.replace('kubernetes:', ''));
    this.apiManager = new ApiManager(
      idToken,
      tokenType,
      parsedToken.preferred_username,
      userGroups,
      parsedToken.namespace[0]
    );
    this.state = {
      name: parsedToken.name,
      registryName: '',
      imageList: new Map(),
      templateLabs: new Map(),
      instanceLabs: new Map(),
      templateLabsAdmin: new Map(),
      instanceLabsAdmin: new Map(),
      adminGroups,
      selectedTemplate: { name: null, namespace: null },
      selectedInstance: null,
      events: '',
      statusHidden: true,
      adminHidden: true
    };
    this.retriveImageList();
    this.retrieveCRDtemplates();
    this.retrieveCRDinstances()
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        /* Start watching for namespaced events */
        this.apiManager.startWatching(this.notifyEvent);

        /* Start watching for admin namespaces events if any */
        if (adminGroups.length > 0) {
          this.apiManager.startWatching(this.notifyEventAdmin, {
            labelSelector: `template-namespace in (${adminGroups.join()})`
          });
        }

        /* @@@@@@@@@@@ TO BE USED ONLY IF WATCHER IS BROKEN
                        this.retrieveCRDinstanceStatus();
                        setInterval(() => {this.retrieveCRDinstanceStatus()}, 10000);
                        */
      });
  }

  /**
   * Private function to retrieve all CRD templates available
   */
  retrieveCRDtemplates() {
    const { templateLabs, templateLabsAdmin, adminGroups } = this.state;
    this.apiManager
      .getCRDtemplates()
      .then(res => {
        const newMap = templateLabs;
        const newMapAdmin = templateLabsAdmin;
        res.forEach(x => {
          if (x) {
            newMap.set(x.course, x.labs);
            if (adminGroups.includes(x.course))
              newMapAdmin.set(x.course, x.labs);
          }
        });
        this.setState({ templateLabs: newMap, templateLabsAdmin: newMapAdmin });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Private function to retrieve all CRD instances running
   */
  retrieveCRDinstances() {
    const { instanceLabs } = this.state;
    return this.apiManager
      .getCRDinstances()
      .then(nodesResponse => {
        const nodes = nodesResponse.body.items;
        const newMap = instanceLabs;
        nodes.forEach(x => {
          if (!newMap.has(x.metadata.name)) {
            newMap.set(x.metadata.name, { status: 0, url: null });
          }
        });
        this.setState({ instanceLabs: newMap });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Function to start and create a CRD instance using the actual selected one
   */
  startCRDinstance() {
    const { selectedTemplate, instanceLabs } = this.state;
    if (!selectedTemplate.name) {
      Toastr.info('Please select a lab before starting it');
      return;
    }
    if (instanceLabs.has(selectedTemplate.name)) {
      Toastr.info(`The \`${selectedTemplate.name}\` lab is already running`);
      return;
    }
    this.apiManager
      .createCRDinstance(selectedTemplate.name, selectedTemplate.namespace)
      .then(response => {
        Toastr.success(`Successfully started lab \`${selectedTemplate.name}\``);
        const newMap = instanceLabs;
        newMap.set(response.body.metadata.name, { status: 0, url: null });
        this.setState({ instanceLabs: newMap });
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        this.changeSelectedCRDtemplate(null, null);
      });
  }

  /**
   * Function to start and create a CRD template
   */
  createCRDtemplate(namespace, labNumber, description, cpu, memory, image) {
    const { templateLabs, templateLabsAdmin } = this.state;
    this.apiManager
      .createCRDtemplate(namespace, labNumber, description, cpu, memory, image)
      .then(response => {
        Toastr.success(`Successfully create template \`${description}\``);
        const newMap = templateLabs;
        newMap.set(response.body.metadata.name, { status: 0, url: null });
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        templateLabs.clear();
        templateLabsAdmin.clear();
        this.retrieveCRDtemplates();
      });
  }

  /**
   * Function to stop and delete the current selected CRD instance
   */
  stopCRDinstance() {
    const { selectedInstance, instanceLabs } = this.state;
    if (!selectedInstance) {
      Toastr.info('No lab to stop has been selected');
      return;
    }
    if (!instanceLabs.has(selectedInstance)) {
      Toastr.info(`The \`${selectedInstance}\` lab is not running`);
      return;
    }
    this.apiManager
      .deleteCRDinstance(selectedInstance)
      .then(() => {
        Toastr.success(`Successfully stopped \`${selectedInstance}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        const newMap = instanceLabs;
        newMap.delete(selectedInstance);
        this.setState({ instanceLabs: newMap });
        this.changeSelectedCRDinstance(null);
      });
  }

  deleteCRDtemplate() {
    const { selectedTemplate, templateLabsAdmin, templateLabs } = this.state;
    if (!selectedTemplate) {
      Toastr.info('No template to delete has been selected');
      return;
    }
    if (!templateLabsAdmin.has(selectedTemplate.namespace)) {
      Toastr.info(
        `The \`${selectedTemplate.name} template is not managed by you`
      );
      return;
    }

    this.apiManager
      .deleteCRDtemplate(selectedTemplate.namespace, selectedTemplate.name)
      .then(() => {
        Toastr.success(`Successfully deleted \`${selectedTemplate.name}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        templateLabs.clear();
        templateLabsAdmin.clear();
        this.retrieveCRDtemplates();
      });
  }

  /**
   * Function to stop and delete the current selected CRD instance
   */
  stopCRDinstanceAdmin() {
    const { selectedInstance, instanceLabsAdmin } = this.state;
    if (!selectedInstance) {
      Toastr.info('No lab to stop has been selected');
      return;
    }
    if (!instanceLabsAdmin.has(selectedInstance)) {
      Toastr.info(`The \`${selectedInstance}\` lab is not running`);
      return;
    }
    this.apiManager
      .deleteCRDinstance(selectedInstance)
      .then(() => {
        Toastr.success(`Successfully stopped \`${selectedInstance}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        const newMap = instanceLabsAdmin;
        newMap.delete(selectedInstance);
        this.setState({ instanceLabsAdmin: newMap });
        this.changeSelectedCRDinstance(null);
      });
  }

  /**
   * Function to perform logout when session with cluster expires
   */
  logoutInterval() {
    setInterval(() => {
      const { logout } = this.props;
      /* A reload probably is sufficient to re-authN the token */
      logout();
    }, 2000);
  }

  /**
   * Function to connect to the VM of the actual user selected CRD instance
   */
  connect() {
    const { selectedInstance, instanceLabs } = this.state;
    if (!selectedInstance) {
      Toastr.info('No running lab selected to connect to');
      return;
    }
    if (!instanceLabs.has(selectedInstance)) {
      Toastr.info(`The lab \`${selectedInstance}\` is not running`);
      return;
    }
    switch (instanceLabs.get(selectedInstance).status) {
      case 1:
        window.open(instanceLabs.get(selectedInstance).url);
        break;
      case 0:
        Toastr.info(`The lab \`${selectedInstance}\` is still starting`);
        break;
      default:
        Toastr.info(
          `An error has occurred with the lab \`${selectedInstance}\``
        );
        break;
    }

    this.changeSelectedCRDinstance(null);
  }

  /**
   * Function to connect to the VM of the actual admin selected CRD instance
   */
  connectAdmin() {
    const { selectedInstance, instanceLabsAdmin } = this.state;
    if (!selectedInstance) {
      Toastr.info('No running lab selected to connect to');
      return;
    }
    if (!instanceLabsAdmin.has(selectedInstance)) {
      Toastr.info(`The lab \`${selectedInstance}\` is not running`);
      return;
    }
    switch (instanceLabsAdmin.get(selectedInstance).status) {
      case 1:
        window.open(instanceLabsAdmin.get(selectedInstance).url);
        break;
      case 0:
        Toastr.info(`The lab \`${selectedInstance}\` is still starting`);
        break;
      default:
        Toastr.info(
          `An error has occurred with the lab \`${selectedInstance}\``
        );
        break;
    }

    this.changeSelectedCRDinstance(null);
  }

  /**
   * * @@@@ UNUSED (since watcher has been patched and works)
   *
   * Function to retrieve all CRD instances status
   */
  retrieveCRDinstanceStatus() {
    const { instanceLabs, events } = this.state;
    const keys = Array.from(instanceLabs.keys());
    keys.forEach(lab => {
      this.apiManager
        .getCRDstatus(lab)
        .then(response => {
          if (response.body.status && response.body.status.phase) {
            const msg = `[${response.body.metadata.creationTimestamp}] ${lab} => ${response.body.status.phase}`;
            const newMap = instanceLabs;
            if (response.body.status.phase.match(/Fail|Not/g)) {
              /* Object creation failed */
              newMap.set(lab, { url: null, status: -1 });
            } else if (response.body.status.phase.match(/VmiReady/g)) {
              /* Object creation succeeded */
              newMap.set(lab, { url: response.body.status.url, status: 1 });
            }
            this.setState({
              instanceLabs: newMap,
              events: `${msg}\n${events}`
            });
          }
        })
        .catch(error => {
          this.handleErrors(error);
        });
    });
  }

  /**
   *Function to notify a Kubernetes Event related to your user resources
   * @param type the type of the event
   * @param object the object of the event
   */
  notifyEvent(type, object) {
    const { instanceLabs, events } = this.state;
    /* TODO: intercept 403 and redirect to logout */
    if (!type) {
      /* Watch session ended, restart it */
      this.apiManager.startWatching(this.notifyEvent);
      this.setState({ events: '' });
      return;
    }
    if (object && object.status) {
      const msg = `[${object.metadata.creationTimestamp}] ${object.metadata.name} {type: ${type}, status: ${object.status.phase}}`;
      const newMap = instanceLabs;
      if (object.status.phase.match(/Fail|Not/g)) {
        /* Object creation failed */
        newMap.set(object.metadata.name, { url: null, status: -1 });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /* Object creation succeeded */
        newMap.set(object.metadata.name, { url: object.status.url, status: 1 });
      } else if (type === 'DELETED') {
        newMap.delete(object.metadata.name);
      }
      this.setState({
        instanceLabs: newMap,
        events: `${msg}\n${events}`
      });
    }
  }

  /**
   *Function to notify a Kubernetes Event related to your admin resources
   * @param type the type of the event
   * @param object the object of the event
   */
  notifyEventAdmin(type, object) {
    const { adminGroups, instanceLabsAdmin } = this.state;
    /* TODO: intercept 403 and redirect to logout */
    if (!type) {
      /* Watch session ended, restart it */
      this.apiManager.startWatching(this.notifyEventAdmin, {
        labelSelector: `template-namespace in (${adminGroups.join()})`
      });
      return;
    }
    if (object && object.status) {
      const newMap = instanceLabsAdmin;
      if (object.status.phase.match(/Fail|Not/g)) {
        /* Object creation failed */
        newMap.set(object.metadata.name, { url: null, status: -1 });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /* Object creation succeeded */
        newMap.set(object.metadata.name, { url: object.status.url, status: 1 });
      } else if (type === 'DELETED') {
        newMap.delete(object.metadata.name);
      } else if (
        (type === 'ADDED' || type === 'MODIFIED') &&
        !newMap.has(object.metadata.name)
      ) {
        newMap.set(object.metadata.name, { url: null, status: 0 });
      }
      this.setState({
        instanceLabsAdmin: newMap
      });
    }
  }

  /**
   * Function to change the user selected CRD template
   * @param name the name/label of the new one
   * @param namespace the namespace in which the template should be retrieved
   */
  changeSelectedCRDtemplate(name, namespace) {
    this.setState({
      selectedTemplate: { name, namespace }
    });
  }

  /**
   * Function to change the user selected CRD instance
   * @param name the name/label of the new one
   */
  changeSelectedCRDinstance(name) {
    this.setState({ selectedInstance: name });
  }

  /**
   * Function to handle all errors
   * @param error the error message received
   */
  handleErrors(error) {
    let msg = '';
    // next eslint-disable is because the k8s_library uses the dash in their implementation
    // eslint-disable-next-line no-underscore-dangle
    switch (error.response._fetchResponse.status) {
      case 401:
        msg += 'Forbidden, something in the ticket renewal failed';
        this.logoutInterval();
        break;
      case 403:
        msg +=
          'It seems you do not have the right permissions to perform this operation';
        break;
      case 404:
        msg += 'Resource not found, probably you have already destroyed it';
        break;
      case 409:
        msg += 'The resource is already present';
        break;
      default:
        // next eslint-disable is because the k8s_library uses the dash in their implementation
        // eslint-disable-next-line no-underscore-dangle
        msg += `An error occurred(${error.response._fetchResponse.status}), please login again`;
        this.logoutInterval();
    }
    Toastr.error(msg);
  }

  /**
   * Private function to retrieve all CRD instances running
   */
  retriveImageList() {
    const { imageList } = this.state;
    this.apiManager
      .retrieveImageList()
      .then(nodesResponse => {
        const newMap = imageList;
        this.setState({ registryName: nodesResponse.body.spec.registryName });
        nodesResponse.body.spec.images.map(x => {
          newMap.set(x.name, x.versions);
          return x.name;
        });
        this.setState({ imageList: newMap });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Function to render this component,
   * It automatically updates every new change in the state variable
   * @returns the component to be drawn
   */
  render() {
    const { logout } = this.props;
    const {
      name,
      adminHidden,
      adminGroups,
      registryName,
      imageList,
      templateLabsAdmin,
      instanceLabsAdmin,
      templateLabs,
      instanceLabs,
      events,
      statusHidden
    } = this.state;
    return (
      <div id="body" style={{ height: '100%', background: '#fafafa' }}>
        <Header
          logged
          logout={logout}
          name={name}
          adminHidden={adminHidden}
          renderAdminBtn={adminGroups.length > 0}
          switchAdminView={() => this.setState({ adminHidden: !adminHidden })}
        />
        <Body
          registryName={registryName}
          retriveImageList={imageList}
          adminGroups={adminGroups}
          templateLabsAdmin={templateLabsAdmin}
          instanceLabsAdmin={instanceLabsAdmin}
          templateLabs={templateLabs}
          funcNewTemplate={this.createCRDtemplate}
          instanceLabs={instanceLabs}
          funcTemplate={this.changeSelectedCRDtemplate}
          funcInstance={this.changeSelectedCRDinstance}
          start={this.startCRDinstance}
          deleteLabTemplate={this.deleteCRDtemplate}
          connect={this.connect}
          connectAdmin={this.connectAdmin}
          stop={this.stopCRDinstance}
          stopAdmin={this.stopCRDinstanceAdmin}
          events={events}
          showStatus={() => this.setState({ statusHidden: !statusHidden })}
          hidden={statusHidden}
          adminHidden={adminHidden}
        />
        <Footer />
      </div>
    );
  }
}
