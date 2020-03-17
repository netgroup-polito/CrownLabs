import React from 'react';
import Login from './Login';
import Test from './Test';
import Home from './Home';
import UserView from './UserView';
import {BrowserRouter as Router, Route, Switch} from 'react-router-dom';
import './App.css';

function App() {
  return (
    <Router>
        <Switch>
          <Route exact path="/">
            <Home />
          </Route>
          <Route path="/login">
            <Login />
          </Route>
          <Route path="/userview">
            <UserView />
          </Route>
          <Route path="/test">
            <Test />
          </Route>
        </Switch>
    </Router>
  );
}

export default App;
