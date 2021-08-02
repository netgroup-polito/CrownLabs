import { useContext, useEffect, useState } from 'react';
import './App.css';
import { PUBLIC_URL } from './env';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { AuthContext, logout } from './contexts/AuthContext';
import { Button } from 'antd';

function App() {
  const { isLoggedIn, userId, token } = useContext(AuthContext);
  const [instances, setInstances] = useState<string[] | undefined>(undefined);
  useEffect(() => {
    if (userId) {
      fetch(
        `https://apiserver.crownlabs.polito.it/apis/crownlabs.polito.it/v1alpha2/namespaces/tenant-${userId.replaceAll(
          '.',
          '-'
        )}/instances`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      )
        .then(res => res.json())
        .then(body => {
          if (body.items) {
            setInstances(body.items.map((item: any) => item.metadata.name));
          }
        })
        .catch(err => {
          console.error('ERROR WHEN GETTING INSTANCES', err);
        });
    }
  }, [userId, token]);
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
          <Route path="/" exact>
            <div className="p-10 m-10 flex flex-col items-center">
              {isLoggedIn && (
                <>
                  {instances?.length === 0
                    ? 'You have no active instances at the moment'
                    : 'Your instances on CrownLabs:'}
                </>
              )}
              {instances?.map(instance => (
                <h6 style={{ fontSize: '1.5rem' }}>{instance}</h6>
              ))}
              {isLoggedIn && (
                <>
                  <Button
                    onClick={() => {
                      logout();
                    }}
                  >
                    LOGOUT
                  </Button>
                </>
              )}
            </div>
          </Route>
        </Switch>
      </BrowserRouter>
    </div>
  );
}

export default App;
