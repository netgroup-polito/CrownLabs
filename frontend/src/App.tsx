import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { useContext } from 'react';
import './App.css';
import {
  VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX,
  VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM,
  VITE_APP_CROWNLABS_GRAFANA_DASHBOARD_URL,
} from './env';
import { TenantContext } from './contexts/TenantContext';
import { LinkPosition } from './utils';
import AppLayout from './components/common/AppLayout';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import SSHTerminal from './components/activePage/SSHTerminal/SSHTerminal';
import DriveView from './components/activePage/DriveView';
import { VITE_APP_MYDRIVE_WORKSPACE_NAME } from './env';
import TenantPage from './components/tenants/TenantPage';
import TenantListPage from './components/tenants/TenantListPage';

function App() {
  const { data: tenantData } = useContext(TenantContext);

  // Check if user has access to utilities workspace
  const hasUtilitiesAccess = Boolean(
    tenantData?.tenant?.spec?.workspaces?.some(
      ws => ws?.name === VITE_APP_MYDRIVE_WORKSPACE_NAME,
    ),
  );

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  // Early return if tenant namespace not available
  if (!tenantNs) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
        }}
      >
        <div>Loading tenant information...</div>
      </div>
    );
  }

  return (
    <AppLayout
      TooltipButtonLink={
        `${VITE_APP_CROWNLABS_GRAFANA_DASHBOARD_URL}?&var-namespace=` +
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
        ...(hasUtilitiesAccess
          ? [
              {
                route: {
                  name: 'Drive',
                  path: '/drive',
                },
                content: <DriveView key="/drive" />,
                linkPosition: LinkPosition.NavbarButton,
              },
            ]
          : []),
        {
          route: { name: 'Users', path: '/tenants' },
          content: <TenantListPage />,
          linkPosition: LinkPosition.NavbarButton,
          requiredGroups: [
            `${VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX}:${VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM}`,
          ],
        },
        {
          route: { name: 'Tenant', path: '/tenants/:tenantId' },
          content: <TenantPage />,
          linkPosition: LinkPosition.Hidden,
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
            path: '/instance/:namespace/:VMname/:environment/ssh',
          },
          content: (
            <SSHTerminal key="/instance/:namespace/:VMname/:environment/ssh" />
          ),
          linkPosition: LinkPosition.WebSSH,
        },
      ]}
    />
  );
}

export default App;
