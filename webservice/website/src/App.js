import React from 'react';
import Home from './Home';
import UserView from './UserView';
import {BrowserRouter as Router, Redirect, Route, Switch} from 'react-router-dom';
import './App.css';
import Authenticator from "./services/Authenticator";
import {Container} from "react-bootstrap";

/**
 * Dumb function to manage the callback. It renders nothing, just call the function passed as props
 * @param props the function to be executed
 */
function CallBackHandler(props) {
    props.func();
    return (<div/>);
}

/**
 * The main class of this project.
 */
export class App extends React.Component {
    constructor(props) {
        super(props);
        /*It hash an authenticator which will be used by all the subclasses*/
        this.authManager = new Authenticator();
        /*The state is composed (by now) by a logged (true|false) variable. The sessionStage is checked since the OIDC library
        * stores there the token if the user is logged */
        this.state = {logged: !!sessionStorage.length};
        this.authManager.manager.startSilentRenew();
        this.authManager.manager.events.addUserLoaded(user => {
            if (user && !this.state.logged) {
                this.setState({logged: true});
            }
        });
    }

    render() {
        /*Maybe one of the best solution would be renaming the UserView variable into MainWindow and delegating to it the rendering
        * of the different pages (user unprivileged, admin or professor)
        * PLEASE REMEMBER, if you change on of these path names (ex /userview into /mainview) you have to modify the NGINX conf of our Ingress*/
        return (
            <Container className="col-9 p-0" style={{backgroundColor: '#FCF6F5FF'}}>
                    <Router>
                        <Switch>
                            <Route exact path="/">
                                <Home login={this.authManager.login}/>
                            </Route>
                            <Route path="/userview" render={() => (
                                this.state.logged ?
                                    <UserView logout={this.authManager.logout}/> :
                                    <Redirect to="/"/>
                            )}/>
                            <Route path="/callback" render={() => (
                                this.state.logged ? <Redirect to="/userview"/> : <CallBackHandler func={this.authManager.completeLogin}/>
                            )}/>
                            <Route path="*">
                                <Redirect to="/userview"/>
                            </Route>
                        </Switch>
                    </Router>
            </Container>
        );
    }
}
