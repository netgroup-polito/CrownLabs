import React from 'react';
import CssBaseline from '@material-ui/core/CssBaseline';
import {
  BrowserRouter as Router,
  Redirect,
  Route,
  Switch
} from 'react-router-dom';
import UserLogic from './UserLogic';
import './App.css';
import Authenticator from './services/Authenticator';
import CallBackHandler from './components/CallBackHandler';
/**
 * The main class of this project.
 */
export default class App extends React.Component {
  /* Unique authN among the entire application;
   * childKey used to redraw UserLogin every token refreshed */
  constructor(props) {
    super(props);
    this.authManager = new Authenticator();
    this.childKey = 0;
    /* Check if previously logged */
    const retrievedSessionToken = JSON.parse(
      sessionStorage.getItem(`oidc.user:${OIDC_PROVIDER_URL}:${OIDC_CLIENT_ID}`)
    );
    if (retrievedSessionToken) {
      this.state = {
        logged: true,
        idToken: retrievedSessionToken.id_token,
        tokenType: retrievedSessionToken.token_type || 'Bearer'
      };
    } else {
      this.state = { logged: false, idToken: null, tokenType: null };
    }
    this.authManager.manager.events.addUserLoaded(user => {
      const { logged } = this.state;
      if (user && !logged) {
        this.setState({
          logged: true,
          idToken: user.id_token,
          tokenType: user.token_type || 'Bearer'
        });
      }
    });
    this.authManager.manager.events.addAccessTokenExpiring(() => {
      this.authManager.manager.signinSilent().then(user => {
        this.childKey += 1;
        this.setState({
          logged: true,
          idToken: user.id_token,
          tokenType: user.token_type || 'Bearer'
        });
      });
    });
  }

  render() {
    return (
      <>
        <CssBaseline />
        <Router>
          <Switch>
            <Route
              exact
              path="/login"
              render={() => {
                this.authManager.login();
              }}
            />
            <Route
              path="/userview"
              render={() => {
                const { logged, idToken, tokenType } = this.state;
                return logged ? (
                  <UserLogic
                    key={this.childKey}
                    idToken={idToken}
                    tokenType={tokenType}
                    logout={this.authManager.logout}
                  />
                ) : (
                  <Redirect to="/login" />
                );
              }}
            />
            <Route
              path="/callback"
              render={() => {
                const { logged } = this.state;
                return logged ? (
                  <Redirect to="/userview" />
                ) : (
                  <CallBackHandler func={this.authManager.completeLogin} />
                );
              }}
            />
            <Route path="*">
              <Redirect to="/login" />
            </Route>
          </Switch>
        </Router>
      </>
    );
  }
}
