import './App.css';
import { PUBLIC_URL } from './env';
import { BrowserRouter, NavLink, Route, Switch } from 'react-router-dom';
import { Layout, Button, Space, Switch as SwitchToggle } from 'antd';
import { useState } from 'react';
import Dashboard from './components/workspaces/Dashboard';
import { Auth } from './components/workspaces/auth';
import { ModalExit } from './components/workspaces/Modal';

const { Header, Footer, Content } = Layout;

function App() {
  const [auth, setAuth] = useState(false);
  const [showExitModal, setshowExitModal] = useState(false);

  return (
    <BrowserRouter basename={PUBLIC_URL}>
      <Layout className="h-screen">
        <Header className="flex justify-center align-center">
          <Space size="small">
            <NavLink exact={true} to="/">
              <Button type="primary" shape="round" size={'large'}>
                Dashboard
              </Button>
            </NavLink>
            <NavLink to="/active">
              <Button type="primary" shape="round" size={'large'}>
                Active
              </Button>
            </NavLink>
            <NavLink to="/account">
              <Button type="primary" shape="round" size={'large'}>
                Account
              </Button>
            </NavLink>
            <SwitchToggle checked={auth} onChange={() => setAuth(!auth)} />
          </Space>
        </Header>
        <Switch>
          <Auth.Provider value={auth}>
            <Route path="/active"></Route>
            <Route path="/account"></Route>
            <Route path="/" exact>
              <Content className="pt-5">
                <Dashboard />
                <ModalExit
                  showmodal={showExitModal}
                  setshowmodal={setshowExitModal}
                />
              </Content>
            </Route>
          </Auth.Provider>
        </Switch>
        <Footer className="flex justify-center items-center">Footer</Footer>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
