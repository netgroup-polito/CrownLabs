import './App.css';
import AppLayout from './components/common/AppLayout';
import ThemeContextProvider from './contexts/ThemeContext';
import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { useContext } from 'react';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import { TenantContext } from './graphql-components/tenantContext/TenantContext';
import { LinkPosition } from './utils';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';

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
            content: <DashboardLogic />,
            linkPosition: LinkPosition.NavbarButton,
          },
          {
            route: { name: 'Active', path: '/active' },
            content: <ActiveViewLogic />,
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
            content: <UserPanelLogic />,
            linkPosition: LinkPosition.MenuButton,
          },
        ]}
      />
    </ThemeContextProvider>
  );
}

export default App;
