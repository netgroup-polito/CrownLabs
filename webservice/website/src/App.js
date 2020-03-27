import React from 'react';
import Home from './views/Home';
import UserLogic from './UserLogic';
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
    /*Unique authN among the entire application;
    * childKey used to redraw UserLogin every token refreshed*/
    constructor(props) {
        super(props);
        this.authManager = new Authenticator();
        this.childKey = 0;
        /*Check if previously logged*/
        let retrievedSessionToken = JSON.parse(sessionStorage.getItem('oidc.user:' + OIDC_PROVIDER_URL + ":" + OIDC_CLIENT_ID));
        if (retrievedSessionToken) {
            this.state = {
                logged: true,
                id_token: retrievedSessionToken.id_token,
                token_type: retrievedSessionToken.token_type || "Bearer"
            };
        } else {
            this.state = {logged: false, id_token: null, token_type: null};
        }
        this.authManager.manager.events.addUserLoaded(user => {
            if (user && !this.state.logged) {
                this.setState({logged: true, id_token: user.id_token, token_type: user.token_type || "Bearer"});
            }
        });
        this.authManager.manager.events.addAccessTokenExpiring(() => {
            this.authManager.manager.signinSilent()
                .then(user => {
                    this.childKey++;
                    this.setState({logged: true, id_token: user.id_token, token_type: user.token_type || "Bearer"});
                });
        });
    }

    render() {
        return (
            <Container className="col-9 p-0" style={{backgroundColor: '#FCF6F5FF'}}>
                <Router>
                    <Switch>
                        <Route exact path="/">
                            <Home login={this.authManager.login}/>
                        </Route>
                        <Route path="/userview" render={() => (
                            this.state.logged ?
                                <UserLogic key={this.childKey} id_token={this.state.id_token}
                                           token_type={this.state.token_type} logout={this.authManager.logout}/> :
                                <Redirect to="/"/>
                        )}/>
                        <Route path="/callback" render={() => (
                            this.state.logged ? <Redirect to="/userview"/> :
                                <CallBackHandler func={this.authManager.completeLogin}/>
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
