import React from 'react';
import Login from './Login';
import Home from './Home';
import './App.css';
import CallBackHandler from "./CallBackHandler";
import Authenticator from "./Authenticator";
import {
  BrowserRouter as Router,
  Switch,
  Route
} from "react-router-dom";

const myAuth = new Authenticator();

export {myAuth as AuthenticatorInstance};

export function App() {
  return (
      <Router>
        <div>
          {/* A <Switch> looks through its children <Route>s and
            renders the first one that matches the current URL. */}
          <Switch>
            <Route path="/callback">
              <CallBackHandler />
            </Route>
            <Route path="/home">
              <Home />
            </Route>
            <Route path="/">
              <Login />
            </Route>
          </Switch>
        </div>
      </Router>
  );
}