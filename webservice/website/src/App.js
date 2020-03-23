import React from 'react';
import Home from './Home';
import UserView from './UserView';
import {BrowserRouter as Router, Redirect, Route, Switch} from 'react-router-dom';
import './App.css';
import Authenticator from "./services/Authenticator";
import {Container} from "react-bootstrap";

function CallBackHandler(props) {
    props.func();
    return (<div/>);
}

export class App extends React.Component {
    constructor(props) {
        super(props);
        this.authManager = new Authenticator();
        this.state = {logged: !!sessionStorage.length};
        this.authManager.manager.startSilentRenew();
        this.authManager.manager.events.addUserLoaded(user => {
            if (user && !this.state.logged) {
                this.setState({logged: true});
            }
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
