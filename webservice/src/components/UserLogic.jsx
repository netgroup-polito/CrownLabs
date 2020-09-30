import React from 'react';
import Toastr from 'toastr';
import ApiManager from '../services/ApiManager';
import 'toastr/build/toastr.min.css';
import { parseJWTtoken, checkToken } from '../helpers';
import Main from './Main';
/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserLogic extends React.Component {
  /* The State variable contains:
   * - all lab templates as a Map: (course_group => Array of available templates for that course)
   * - all lab instances as a Map: (instance_name => URL if running, null otherwise)
   * - all ADMIN lab templates as a Map: (course_group => Array of available templates for that course)
   * - all ADMIN lab instances as a Map: (instance_name => URL if running, null otherwise)
   * - boolean variable whether to show the status info area
   * - isStudentView whether to render or not the admin page (changed by the button in the StudentView IF adminGroups is not false)
   * */

  constructor(props) {
    super(props);
    const { idToken, tokenType } = props;
    this.connect = this.connect.bind(this);
    this.retrieveImageList = this.retrieveImageList.bind(this);
    this.startCRDinstance = this.startCRDinstance.bind(this);
    this.stopCRDinstance = this.stopCRDinstance.bind(this);
    this.deleteCRDtemplate = this.deleteCRDtemplate.bind(this);
    this.createCRDtemplate = this.createCRDtemplate.bind(this);
    this.stopCRDinstanceAdmin = this.stopCRDinstanceAdmin.bind(this);
    this.notifyEvent = this.notifyEvent.bind(this);
    this.connectAdmin = this.connectAdmin.bind(this);
    this.notifyEventAdmin = this.notifyEventAdmin.bind(this);
    this.changeAdminView = this.changeAdminView.bind(this);

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
    let prevIsStudentView = JSON.parse(localStorage.getItem('isStudentView'));
    if (prevIsStudentView === null) prevIsStudentView = true;
    this.state = {
      name: parsedToken.name,
      registryName: '',
      imageList: new Map(),
      templateLabs: new Map(),
      instanceLabs: new Map(),
      templateLabsAdmin: new Map(),
      instanceLabsAdmin: new Map(),
      adminGroups,
      isStudentView: prevIsStudentView,
      descriptions: {},
      instanceTypes: {}
    };
    this.retrieveImageList();
    this.retrieveCRDtemplates().then(() => {
      /* Start watching for namespaced events */
      this.apiManager.startWatching(this.notifyEvent);
      /* Start watching for admin namespaces events if any */
      if (adminGroups.length > 0) {
        this.apiManager.startWatching(this.notifyEventAdmin, {
          labelSelector: `template-namespace in (${adminGroups.join()})`
        });
      }
    });
  }

  /**
   * Private function to retrieve all CRD templates available
   */
  retrieveCRDtemplates() {
    const {
      templateLabs,
      templateLabsAdmin,
      adminGroups,
      descriptions,
      instanceTypes
    } = this.state;
    return this.apiManager
      .getCRDtemplates()
      .then(res => {
        const newMap = templateLabs;
        const newMapAdmin = templateLabsAdmin;
        const newDescriptions = { ...descriptions };
        const newInstanceTypes = { ...instanceTypes };
        res.forEach(x => {
          if (x) {
            newMap.set(
              x.course,
              x.labs.map(({ name, description, type }) => ({
                labName: name,
                description,
                type
              }))
            );
            if (adminGroups.includes(x.course))
              newMapAdmin.set(
                x.course,
                x.labs.map(({ name, description, type }) => ({
                  labName: name,
                  description,
                  type
                }))
              );
            x.labs.forEach(({ name, description, type }) => {
              newDescriptions[name] = description;
              newInstanceTypes[name] = type;
            });
          }
        });
        this.setState({
          templateLabs: newMap,
          templateLabsAdmin: newMapAdmin,
          descriptions: newDescriptions,
          instanceTypes: newInstanceTypes
        });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Function to start and create a CRD instance
   */
  startCRDinstance(templateName, course) {
    const { instanceLabs, descriptions, instanceTypes } = this.state;
    if (instanceLabs.has(templateName)) {
      Toastr.info(`The \`${templateName}\` lab is already running`);
      return;
    }
    this.apiManager
      .createCRDinstance(templateName, course)
      .then(response => {
        Toastr.success(`Successfully started lab \`${templateName}\``);
        const newMap = instanceLabs;
        newMap.set(response.body.metadata.name, {
          status: 0,
          url: null,
          ip: '',
          creationTime: undefined,
          description: descriptions[templateName],
          type: instanceTypes[templateName]
        });
        this.setState({ instanceLabs: newMap });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Function to start and create a CRD template
   */
  createCRDtemplate(
    namespace,
    labNumber,
    description,
    cpu,
    memory,
    image,
    type
  ) {
    this.apiManager
      .createCRDtemplate(
        namespace,
        labNumber,
        description,
        cpu,
        memory,
        image,
        type
      )
      .then(() => {
        Toastr.success(`Successfully created template \`${description}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        this.retrieveCRDtemplates();
      });
  }

  /**
   * Function to stop and delete an instance
   */
  stopCRDinstance(instanceName) {
    const { instanceLabs } = this.state;
    if (!instanceLabs.has(instanceName)) {
      Toastr.info(`The \`${instanceName}\` lab is not running`);
      return;
    }
    this.apiManager
      .deleteCRDinstance(instanceName)
      .then(() => {
        Toastr.success(`Successfully stopped \`${instanceName}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        const newMap = instanceLabs;
        newMap.delete(instanceName);
        this.setState({ instanceLabs: newMap });
      });
  }

  deleteCRDtemplate(templateName, course) {
    const { templateLabsAdmin, templateLabs } = this.state;
    if (!templateLabsAdmin.has(course)) {
      Toastr.info(`The \`${templateName} template is not managed by you`);
      return;
    }

    this.apiManager
      .deleteCRDtemplate(course, templateName)
      .then(() => {
        Toastr.success(`Successfully deleted \`${templateName}\``);
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
   * Function to stop and delete an instance
   */
  stopCRDinstanceAdmin(instanceName) {
    const { instanceLabsAdmin } = this.state;
    if (!instanceLabsAdmin.has(instanceName)) {
      Toastr.info(`The \`${instanceName}\` lab is not running`);
      return;
    }
    this.apiManager
      .deleteCRDinstance(
        instanceName,
        instanceLabsAdmin.get(instanceName).studNamespace
      )
      .then(() => {
        Toastr.success(`Successfully stopped \`${instanceName}\``);
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        const newMap = instanceLabsAdmin;
        newMap.delete(instanceName);
        this.setState({ instanceLabsAdmin: newMap });
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
   * Function to connect to the VM of
   */
  connect(instanceName) {
    const { instanceLabs } = this.state;
    if (!instanceLabs.has(instanceName)) {
      Toastr.info(`The lab \`${instanceName}\` is not running`);
      return;
    }
    switch (instanceLabs.get(instanceName).status) {
      case 1:
        window.open(instanceLabs.get(instanceName).url);
        break;
      case 0:
        Toastr.info(`The lab \`${instanceName}\` is still starting`);
        break;
      default:
        Toastr.info(`An error has occurred with the lab \`${instanceName}\``);
        break;
    }
  }

  /**
   * Function to connect to the VM of admin
   */
  connectAdmin(instanceName) {
    const { instanceLabsAdmin } = this.state;
    if (!instanceLabsAdmin.has(instanceName)) {
      Toastr.info(`The lab \`${instanceName}\` is not running`);
      return;
    }
    switch (instanceLabsAdmin.get(instanceName).status) {
      case 1:
        window.open(instanceLabsAdmin.get(instanceName).url);
        break;
      case 0:
        Toastr.info(`The lab \`${instanceName}\` is still starting`);
        break;
      default:
        Toastr.info(`An error has occurred with the lab \`${instanceName}\``);
        break;
    }
  }

  /**
   *Function to notify a Kubernetes Event related to your user resources
   * @param type the type of the event
   * @param object the object of the event
   */
  notifyEvent(type, object) {
    const { instanceLabs, descriptions, instanceTypes } = this.state;
    /* TODO: intercept 403 and redirect to logout */
    if (!type) {
      /* Watch session ended, restart it */
      this.apiManager.startWatching(this.notifyEvent);
      return;
    }
    if (object && object.status) {
      const newMap = instanceLabs;
      if (object.status.phase.match(/Fail|Not/g)) {
        /* Object creation failed */
        newMap.set(object.metadata.name, {
          url: null,
          status: -1,
          ip: null,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName]
        });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /* Object creation succeeded */
        newMap.set(object.metadata.name, {
          url: object.status.url,
          status: 1,
          ip: object.status.ip,
          creationTime: object.metadata.creationTimestamp,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName]
        });
      } else if (type === 'DELETED') {
        newMap.delete(object.metadata.name);
      } else if (
        (type === 'ADDED' || type === 'MODIFIED') &&
        !newMap.has(object.metadata.name)
      ) {
        newMap.set(object.metadata.name, {
          url: null,
          status: 0,
          ip: null,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName]
        });
      }
      this.setState({
        instanceLabs: newMap
      });
    }
  }

  /**
   *Function to notify a Kubernetes Event related to your admin resources
   * @param type the type of the event
   * @param object the object of the event
   */
  notifyEventAdmin(type, object) {
    const {
      adminGroups,
      instanceLabsAdmin,
      descriptions,
      instanceTypes
    } = this.state;
    /* TODO: intercept 403 and redirect to logout */
    if (!type) {
      /* Watch session ended, restart it */
      this.apiManager.startWatching(this.notifyEventAdmin, {
        labelSelector: `template-namespace in (${adminGroups.join()})`
      });
      return;
    }
    if (object && object.status) {
      const newMap = new Map(instanceLabsAdmin);
      if (object.status.phase.match(/Fail|Not/g)) {
        /* Object creation failed */
        newMap.set(object.metadata.name, {
          url: null,
          status: -1,
          ip: null,
          studNamespace: object.metadata.namespace,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName],
          studentId: object.spec.studentId
        });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /* Object creation succeeded */
        newMap.set(object.metadata.name, {
          url: object.status.url,
          status: 1,
          ip: object.status.ip,
          creationTime: object.metadata.creationTimestamp,
          studNamespace: object.metadata.namespace,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName],
          studentId: object.spec.studentId
        });
      } else if (type === 'DELETED') {
        newMap.delete(object.metadata.name);
      } else if (
        (type === 'ADDED' || type === 'MODIFIED') &&
        !newMap.has(object.metadata.name)
      ) {
        newMap.set(object.metadata.name, {
          url: null,
          status: 0,
          studNamespace: object.metadata.namespace,
          ip: null,
          description: descriptions[object.spec.labTemplateName],
          type: instanceTypes[object.spec.labTemplateName],
          studentId: object.spec.studentId
        });
      }
      this.setState({
        instanceLabsAdmin: newMap
      });
    }
  }

  /**
   * Function to handle all errors
   * @param error the error message received
   */
  handleErrors(error) {
    let msg = '';
    // next eslint-disable is because the k8s_library uses the dash in their implementation
    // eslint-disable-next-line no-underscore-dangle
    if (error.response && error._fetchResponse)
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
          msg += `An error occurred(${
            // next eslint-disable is because the k8s_library uses the dash in their implementation
            // eslint-disable-next-line no-underscore-dangle
            error && error.response._fetchResponse.status
          }), please login again`;
          this.logoutInterval();
      }
    else {
      console.error(error);
      msg += `An error occurred(${
        // next eslint-disable is because the k8s_library uses the dash in their implementation
        // eslint-disable-next-line no-underscore-dangle
        error
      }), please login again`;
      this.logoutInterval();
    }
    Toastr.error(msg);
  }

  /**
   * Private function to retrieve all CRD instances running
   */
  retrieveImageList() {
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

  changeAdminView() {
    const { isStudentView } = this.state;
    localStorage.setItem('isStudentView', JSON.stringify(!isStudentView));
    this.setState({ isStudentView: !isStudentView });
  }

  render() {
    const { logout } = this.props;
    const {
      name,
      isStudentView,
      adminGroups,
      registryName,
      imageList,
      templateLabsAdmin,
      instanceLabsAdmin,
      templateLabs,
      instanceLabs
    } = this.state;
    return (
      <Main
        name={name}
        logout={logout}
        registryName={registryName}
        imageList={imageList}
        adminGroups={adminGroups}
        templateLabsAdmin={templateLabsAdmin}
        instanceLabsAdmin={instanceLabsAdmin}
        templateLabs={templateLabs}
        createTemplate={this.createCRDtemplate}
        instanceLabs={instanceLabs}
        start={this.startCRDinstance}
        deleteLabTemplate={this.deleteCRDtemplate}
        connect={this.connect}
        connectAdmin={this.connectAdmin}
        stop={this.stopCRDinstance}
        stopAdmin={this.stopCRDinstanceAdmin}
        isStudentView={isStudentView}
        changeAdminView={this.changeAdminView}
        createCRDtemplate={this.createCRDtemplate}
        startCRDinstance={this.startCRDinstance}
        deleteCRDtemplate={this.deleteCRDtemplate}
        stopCRDinstance={this.stopCRDinstance}
        stopCRDinstanceAdmin={this.stopCRDinstanceAdmin}
      />
    );
  }
}
