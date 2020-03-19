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
        const authManager = new Authenticator();
        if (localStorage.getItem('isLoggedIn')) {
            this.state = {isLoggedIn: localStorage.getItem('isLoggedIn'), authManager: authManager};
        } else {
            this.state = {isLoggedIn: 'false', authManager: authManager};
        }
        authManager.manager.events.addUserLoaded(user => {
            if (user != null) {
                this.setState({isLoggedIn: 'true'}, () => {
                    localStorage.setItem('isLoggedIn', 'true');
                    localStorage.setItem('token', user.id_token);
                    localStorage.setItem('token_type', user.token_type != null ? user.token_type : "Bearer");
                });
            } else {
                localStorage.setItem('isLoggedIn', 'true');
            }
        });
        authManager.manager.events.addAccessTokenExpired(() => {
            alert("AAAAAAAAAAAAAAAAAAAAAAAa");
        });
        authManager.manager.events.addUserUnloaded(() => {
            localStorage.clear();
            this.setState({isLoggedIn: 'false'});
        })
    }

    render() {
        return (
            <Router>
                <Switch>
                    <Route exact path="/">
                        <Home authManager={this.state.authManager}/>
                    </Route>
                    <Route path="/userview" render={() => (
                        this.state.isLoggedIn === 'true' ?
                            <UserView authManager={this.state.authManager}/> :
                            <Redirect to="/"/>
                    )}/>
                    <Route path="/callback" render={() => (
                        this.state.isLoggedIn === 'true' ? <Redirect to="/userview"/> :
                            <CallBackHandler authManager={this.state.authManager} action={'login'}/>
                    )}/>
                    <Route path="/logout" render={() => (
                        this.state.isLoggedIn === 'true' ?
                            <CallBackHandler authManager={this.state.authManager} action={'logout'}/> :
                            <Redirect to="/"/>
                    )}/>
                    <Route path="*">
                        <Redirect to="/userview"/>
                    </Route>
                </Switch>
            </Router>
        );
    }
}
