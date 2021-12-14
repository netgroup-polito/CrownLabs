import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { useContext } from 'react';
import './App.css';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import AppLayout from './components/common/AppLayout';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import ThemeContextProvider from './contexts/ThemeContext';
import { TenantContext } from './graphql-components/tenantContext/TenantContext';
import { LinkPosition } from './utils';

function App() {
  const { data: tenantData } = useContext(TenantContext);

  return (
    <ThemeContextProvider>
      <AppLayout
        TooltipButtonLink={
          'https://grafana.crownlabs.polito.it/d/BOZGskUGz/personal-overview?&var-namespace=' +
          tenantData?.tenant?.status?.personalNamespace?.name!
        }
        TooltipButtonData={{
          tooltipPlacement: 'left',
          tooltipTitle: 'Statistics',
          icon: (
            <BarChartOutlined
              style={{ fontSize: '22px' }}
              className="flex items-center justify-center "
            />
          ),
          type: 'success',
        }}
        routes={[
          {
            route: { name: 'Dashboard', path: '/' },
            content: <DashboardLogic key="/" />,
            linkPosition: LinkPosition.NavbarButton,
          },
          {
            route: { name: 'Active', path: '/active' },
            content: <ActiveViewLogic key="/active" />,
            linkPosition: LinkPosition.NavbarButton,
          },
          {
            route: { name: 'Drive', path: 'https://crownlabs.polito.it/cloud' },
            linkPosition: LinkPosition.NavbarButton,
          },
          {
            route: {
              name: 'Support',
              path: 'https://support.crownlabs.polito.it/',
            },
            linkPosition: LinkPosition.NavbarButton,
          },
          {
            route: {
              name: 'Manage account',
              path: '/account',
              navbarMenuIcon: <UserOutlined />,
            },
            content: <UserPanelLogic key="/account" />,
            linkPosition: LinkPosition.MenuButton,
          },
        ]}
      />
    </ThemeContextProvider>
  );
}

export default App;
