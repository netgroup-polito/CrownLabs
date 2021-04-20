import './App.css';
import { REACT_APP_CROWNLABS_APISERVER_URL, PUBLIC_URL } from './env';
import { BrowserRouter, Link, Route, Switch } from 'react-router-dom';

function App() {
  return (
    <div
      style={{
        backgroundColor: '#003676',
        color: 'white',
        margin: 'auto',
        width: '100%',
        height: '100%',
        fontSize: '2.7rem',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        textAlign: 'center'
      }}
    >
      <BrowserRouter basename={PUBLIC_URL}>
        <Switch>
          <Route path="/active">
            ACTIVE
            <Link to="/account">to account</Link>
          </Route>
          <Route path="/account">
            ACCOUNT
            <Link to="/active">to active</Link>
          </Route>
          <Route path="/" exact>
            CrownLabs will get a new look! <br /> Apiserver at{' '}
            {REACT_APP_CROWNLABS_APISERVER_URL}
          </Route>
        </Switch>
      </BrowserRouter>
    </div>
  );
}

export default App;
