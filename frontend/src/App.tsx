import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { useContext } from 'react';
import './App.css';
import { TenantContext } from './contexts/TenantContext';
import { LinkPosition } from './utils';
import AppLayout from './components/common/AppLayout';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import SSHTerminal from './components/activePage/SSHTerminal/SSHTerminal';
function App() {
  const { data: tenantData } = useContext(TenantContext);

  return (
    <AppLayout
      TooltipButtonLink={
        'https://grafana.crownlabs.polito.it/d/BOZGskUGz/personal-overview?&var-namespace=' +
        tenantData?.tenant?.status?.personalNamespace?.name
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
        color: 'green',
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
          route: {
            name: 'Drive',
            path: 'https://crownlabs.polito.it/cloud',
          },
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
        {
          route: {
            name: 'Web SSH',
            path: '/ssh/:namespace/:VMname',
          },
          content: <SSHTerminal key="/ssh/:namespace/:VMname" />,
          linkPosition: LinkPosition.Hidden,
        },
      ]}
    />
  );
}

export default App;
