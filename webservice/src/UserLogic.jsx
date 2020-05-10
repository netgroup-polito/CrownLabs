import React, { useState } from 'react';
import Footer from './components/Footer';
import Header from './components/Header';
import ApiManager from './services/ApiManager';
import Toastr from 'toastr';

import 'toastr/build/toastr.min.css';
import Body from './components/Body';

/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserLogic extends React.Component {
  /*The State variable contains:
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
    this.connect = this.connect.bind(this);
    this.changeSelectedCRDtemplate = this.changeSelectedCRDtemplate.bind(this);
    this.changeSelectedCRDinstance = this.changeSelectedCRDinstance.bind(this);
    this.startCRDinstance = this.startCRDinstance.bind(this);
    this.stopCRDinstance = this.stopCRDinstance.bind(this);
    this.createCRDtemplate = this.createCRDtemplate.bind(this);
    this.notifyEvent = this.notifyEvent.bind(this);
    this.connectAdmin = this.connectAdmin.bind(this);
    this.notifyEventAdmin = this.notifyEventAdmin.bind(this);
    let parsedToken = this.parseJWTtoken(this.props.id_token);
    this.theme = false;
    if (!this.checkToken(parsedToken)) {
      this.logoutInterval();
    }
    /*Differentiate the two different kind of group: where the user is admin (professor or PhD) and the one where he is just a student*/
    let adminGroups = parsedToken.groups
      .filter(x => x.match(/kubernetes:\S+admin/g))
      .map(x => x.replace('kubernetes:', '').replace('-admin', ''));
    let userGroups = parsedToken.groups
      .filter(x => x.includes('kubernetes:') && !x.includes('-admin'))
      .map(x => x.replace('kubernetes:', ''));
    this.apiManager = new ApiManager(
      this.props.id_token,
      this.props.token_type,
      parsedToken.preferred_username,
      userGroups,
      parsedToken.namespace[0]
    );
    this.state = {
      name: parsedToken['name'],
      templateLabs: new Map(),
      instanceLabs: new Map(),
      templateLabsAdmin: new Map(),
      instanceLabsAdmin: new Map(),
      adminGroups: adminGroups,
      selectedTemplate: { name: null, namespace: null },
      selectedInstance: null,
      events: '',
      eventsAdmin: '',
      statusHidden: true,
      adminHidden: true
    };
    this.retrieveCRDtemplates();
    this.retrieveCRDinstances()
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        /*Start watching for namespaced events*/
        this.apiManager.startWatching(this.notifyEvent);

        /*Start watching for admin namespaces events if any*/
        if (adminGroups.length > 0) {
          this.apiManager.startWatching(this.notifyEventAdmin, {
            labelSelector: 'template-namespace in (' + adminGroups.join() + ')'
          });
        }

        /* @@@@@@@@@@@ TO BE USED ONLY IF WATCHER IS BROKEN
                        this.retrieveCRDinstanceStatus();
                        setInterval(() => {this.retrieveCRDinstanceStatus()}, 10000);
                        */
      });

    // window._chatlio = window._chatlio || [];
    // !function () {
    //   const t = document.getElementById("chatlio-widget-embed");
    //   if (t && window.ChatlioReact && _chatlio.init)
    //     return void _chatlio.init(t, ChatlioReact);
    //   let e = function (t) { return function () { _chatlio.push([t].concat(arguments))} },
    //       i = ["configure", "identify", "track", "show", "hide", "isShown", "isOnline", "page", "open", "showOrHide"],
    //       a = 0;
    //   for (; a < i.length; a++)
    //     _chatlio[i[a]] || (_chatlio[i[a]] = e(i[a]));
    //   const n = document.createElement("script"), c = document.getElementsByTagName("script")[0];
    //   n.id = "chatlio-widget-embed", n.src = "https://w.chatlio.com/w.chatlio-widget.js",
    //       n.async = !0 ,
    //       n.setAttribute("data-embed-version", "2.3");
    //   n.setAttribute('data-widget-id', '72f8dee1-79f9-48ef-716b-14b677bd57a5');
    //   c.parentNode.insertBefore(n, c);
    // }();
  }

  /**
   * Function to parse a JWT token
   * @param token the token received by keycloak
   * @returns {any} the decrypted token as a JSON object
   */
  parseJWTtoken(token) {
    let base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
    return JSON.parse(
      decodeURIComponent(
        atob(base64)
          .split('')
          .map(function (c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
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
  checkToken(parsed) {
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
   * Private function to retrieve all CRD templates available
   */
  retrieveCRDtemplates() {
    this.apiManager
      .getCRDtemplates()
      .then(res => {
        let newMap = this.state.templateLabs;
        let newMapAdmin = this.state.templateLabsAdmin;
        res.forEach(x => {
          if (x) {
            newMap.set(x.course, x.labs);
            if (this.state.adminGroups.includes(x.course))
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
    return this.apiManager
      .getCRDinstances()
      .then(nodesResponse => {
        const nodes = nodesResponse.body.items;
        let newMap = this.state.instanceLabs;
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
    if (!this.state.selectedTemplate.name) {
      Toastr.info('Please select a lab before starting it');
      return;
    }
    if (this.state.instanceLabs.has(this.state.selectedTemplate.name)) {
      Toastr.info(
        'The `' + this.state.selectedTemplate.name + '` lab is already running'
      );
      return;
    }
    this.apiManager
      .createCRDinstance(
        this.state.selectedTemplate.name,
        this.state.selectedTemplate.namespace
      )
      .then(response => {
        Toastr.success(
          'Successfully started lab `' + this.state.selectedTemplate.name + '`'
        );
        const newMap = this.state.instanceLabs;
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
   * Function to start and create a CRD instance using the actual selected one
   */
  createCRDtemplate(namespace, lab_number, description, cpu, memory, image) {
    this.apiManager
      .createCRDtemplate(namespace, lab_number, description, cpu, memory, image)
      .then(response => {
        Toastr.success('Successfully create template `' + description + '`');
        const newMap = this.state.instanceLabs;
        newMap.set(response.body.metadata.name, { status: 0, url: null });
        this.setState({ instanceLabs: newMap });
      })
      .catch(error => {
        this.handleErrors(error);
      });
  }

  /**
   * Function to stop and delete the current selected CRD instance
   */
  stopCRDinstance() {
    if (!this.state.selectedInstance) {
      Toastr.info('No lab to stop has been selected');
      return;
    }
    if (!this.state.instanceLabs.has(this.state.selectedInstance)) {
      Toastr.info(
        'The `' + this.state.selectedInstance + '` lab is not running'
      );
      return;
    }
    this.apiManager
      .deleteCRDinstance(this.state.selectedInstance)
      .then(() => {
        Toastr.success(
          'Successfully stopped `' + this.state.selectedInstance + '`'
        );
      })
      .catch(error => {
        this.handleErrors(error);
      })
      .finally(() => {
        const newMap = this.state.instanceLabs;
        newMap.delete(this.state.selectedInstance);
        this.setState({ instanceLabs: newMap });
        this.changeSelectedCRDinstance(null);
      });
  }

  /**
   * Function to perform logout when session with cluster expires
   */
  logoutInterval() {
    setInterval(() => {
      /*A reload probably is sufficient to re-authN the token*/
      this.props.logout();
    }, 2000);
  }

  /**
   * Function to connect to the VM of the actual user selected CRD instance
   */
  connect() {
    if (!this.state.selectedInstance) {
      Toastr.info('No running lab selected to connect to');
      return;
    } else if (!this.state.instanceLabs.has(this.state.selectedInstance)) {
      Toastr.info(
        'The lab `' + this.state.selectedInstance + '` is not running'
      );
      return;
    } else {
      switch (this.state.instanceLabs.get(this.state.selectedInstance).status) {
        case 1:
          window.open(
            this.state.instanceLabs.get(this.state.selectedInstance).url
          );
          break;
        case 0:
          Toastr.info(
            'The lab `' + this.state.selectedInstance + '` is still starting'
          );
          break;
        default:
          Toastr.info(
            'An error has occurred with the lab `' +
              this.state.selectedInstance +
              '`'
          );
          break;
      }
    }
    this.changeSelectedCRDinstance(null);
  }

  /**
   * Function to connect to the VM of the actual admin selected CRD instance
   */
  connectAdmin() {
    if (!this.state.selectedInstance) {
      Toastr.info('No running lab selected to connect to');
      return;
    } else if (!this.state.instanceLabsAdmin.has(this.state.selectedInstance)) {
      Toastr.info(
        'The lab `' + this.state.selectedInstance + '` is not running'
      );
      return;
    } else {
      switch (
        this.state.instanceLabsAdmin.get(this.state.selectedInstance).status
      ) {
        case 1:
          window.open(
            this.state.instanceLabsAdmin.get(this.state.selectedInstance).url
          );
          break;
        case 0:
          Toastr.info(
            'The lab `' + this.state.selectedInstance + '` is still starting'
          );
          break;
        default:
          Toastr.info(
            'An error has occurred with the lab `' +
              this.state.selectedInstance +
              '`'
          );
          break;
      }
    }
    this.changeSelectedCRDinstance(null);
  }

  /**
   * * @@@@ UNUSED (since watcher has been patched and works)
   *
   * Function to retrieve all CRD instances status
   */
  retrieveCRDinstanceStatus() {
    const keys = Array.from(this.state.instanceLabs.keys());
    keys.forEach(lab => {
      this.apiManager
        .getCRDstatus(lab)
        .then(response => {
          if (response.body.status && response.body.status.phase) {
            let msg =
              '[' +
              response.body.metadata.creationTimestamp +
              '] ' +
              lab +
              ' => ' +
              response.body.status.phase;
            const newMap = this.state.instanceLabs;
            if (response.body.status.phase.match(/Fail|Not/g)) {
              /*Object creation failed*/
              newMap.set(lab, { url: null, status: -1 });
            } else if (response.body.status.phase.match(/VmiReady/g)) {
              /*Object creation succeeded*/
              newMap.set(lab, { url: response.body.status.url, status: 1 });
            }
            this.setState({
              instanceLabs: newMap,
              events: msg + '\n' + this.state.events
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
    /*TODO: intercept 403 and redirect to logout*/
    if (!type) {
      /*Watch session ended, restart it*/
      this.apiManager.startWatching(this.notifyEvent);
      this.setState({ events: '' });
      return;
    }
    if (object && object.status) {
      let msg =
        '[' +
        object.metadata.creationTimestamp +
        '] ' +
        object.metadata.name +
        ' {type: ' +
        type +
        ', status: ' +
        object.status.phase +
        '}';
      const newMap = this.state.instanceLabs;
      if (object.status.phase.match(/Fail|Not/g)) {
        /*Object creation failed*/
        newMap.set(object.metadata.name, { url: null, status: -1 });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /*Object creation succeeded*/
        newMap.set(object.metadata.name, { url: object.status.url, status: 1 });
      } else if (type === 'DELETED') {
        newMap.delete(object.metadata.name);
      }
      this.setState({
        instanceLabs: newMap,
        events: msg + '\n' + this.state.events
      });
    }
  }

  /**
   *Function to notify a Kubernetes Event related to your admin resources
   * @param type the type of the event
   * @param object the object of the event
   */
  notifyEventAdmin(type, object) {
    /*TODO: intercept 403 and redirect to logout*/
    if (!type) {
      /*Watch session ended, restart it*/
      this.apiManager.startWatching(this.notifyEventAdmin, {
        labelSelector:
          'template-namespace in (' + this.state.adminGroups.join() + ')'
      });
      this.setState({ eventsAdmin: '' });
      return;
    }
    if (object && object.status) {
      let msg =
        '[' +
        object.metadata.creationTimestamp +
        '] ' +
        object.metadata.name +
        ' {type: ' +
        type +
        ', status: ' +
        object.status.phase +
        '}';
      const newMap = this.state.instanceLabsAdmin;
      if (object.status.phase.match(/Fail|Not/g)) {
        /*Object creation failed*/
        newMap.set(object.metadata.name, { url: null, status: -1 });
      } else if (
        object.status.phase.match(/VmiReady/g) &&
        (type === 'ADDED' || type === 'MODIFIED')
      ) {
        /*Object creation succeeded*/
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
        instanceLabsAdmin: newMap,
        eventsAdmin: msg + '\n' + this.state.events
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
      selectedTemplate: { name: name, namespace: namespace }
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
        msg +=
          'An error occurred(' +
          error.response._fetchResponse.status +
          '), please login again';
        this.logoutInterval();
    }
    Toastr.error(msg);
  }

  /**
   * Function to render this component,
   * It automatically updates every new change in the state variable
   * @returns the component to be drawn
   */
  render() {
    return (
      <div id="body" style={{ height: '100%', background: '#fafafa' }}>
        <Header
          logged={true}
          logout={this.props.logout}
          name={this.state.name}
          adminHidden={this.state.adminHidden}
          renderAdminBtn={this.state.adminGroups.length > 0}
          switchAdminView={() =>
            this.setState({ adminHidden: !this.state.adminHidden })
          }
        />
        <Body
          templateLabs={this.state.templateLabs}
          funcNewTemplate={this.createCRDtemplate}
          instanceLabs={this.state.instanceLabs}
          funcTemplate={this.changeSelectedCRDtemplate}
          funcInstance={this.changeSelectedCRDinstance}
          start={this.startCRDinstance}
          connect={this.connect}
          stop={this.stopCRDinstance}
          events={this.state.events}
          showStatus={() =>
            this.setState({ statusHidden: !this.state.statusHidden })
          }
          hidden={this.state.statusHidden}
          adminHidden={this.state.adminHidden}
        />
        <Footer />
      </div>
    );
  }
}
