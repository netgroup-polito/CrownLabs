import './App.css';
import { PUBLIC_URL, REACT_APP_CROWNLABS_APISERVER_URL } from './env';
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
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        textAlign: 'center',
      }}
    >
      <BrowserRouter basename={PUBLIC_URL}>
        <Switch>
          <Route path="/active">
            <div>Active</div>
            <Link to="/">Go Home</Link>
          </Route>
          <Route path="/account">
            <div>Account</div>
            <Link to="/">Go Home</Link>
          </Route>
          <Route path="/" exact>
            <div className="p-10 m-10">
              CrownLabs will get a new look!
              <br /> Apiserver at {REACT_APP_CROWNLABS_APISERVER_URL}{' '}
            </div>

            <Link to="/active">Go to active</Link>
            <Link to="/account">Go to account</Link>
          </Route>
        </Switch>
      </BrowserRouter>
    </div>
  );
}

export default App;
