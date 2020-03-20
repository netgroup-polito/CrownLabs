import React from 'react';
import Home from './Home';
import UserView from './UserView';
import {BrowserRouter as Router, Redirect, Route, Switch} from 'react-router-dom';
import './App.css';
import CallBackHandler from "./CallBackHandler";
import Authenticator from "./services/Authenticator";

export class App extends React.Component {
    constructor(props) {
        super(props);
        this.authManager = new Authenticator();
        this.state = {logged : !!sessionStorage.length};
        this.authManager.manager.events.addUserLoaded(user => {
            if (user && !this.state.logged) {
                this.setState({logged: true});
            }
        });
        this.authManager.manager.events.addUserUnloaded(() => {document.location.href = '/logout';});
    }

    render() {
        return (
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
                        this.state.logged ? <Redirect to="/userview"/> :
                            <CallBackHandler func={this.authManager.completeLogin}/>
                    )}/>
                    <Route path="/logout">
                        <CallBackHandler func={() => {
                            this.setState({logged: false});
                        }}/>
                        <Redirect to="/"/>
                    </Route>
                    <Route path="*">
                        <Redirect to="/userview"/>
                    </Route>
                </Switch>
            </Router>
        );
    }
}
