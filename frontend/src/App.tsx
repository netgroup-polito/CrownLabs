import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { lazy, Suspense, useContext } from 'react';
import './App.css';
import ThemeContextProvider from './contexts/ThemeContextProvider';
import { TenantContext } from './contexts/TenantContext';
import { LinkPosition } from './utils';
import FullPageLoader from './components/common/FullPageLoader';

function App() {
  const { data: tenantData } = useContext(TenantContext);

  const AppLayout = lazy(
    () => import('./components/common/AppLayout/AppLayout'),
  );

  const ActiveViewLogic = lazy(
    () => import('./components/activePage/ActiveViewLogic/ActiveViewLogic'),
  );
  const UserPanelLogic = lazy(
    () => import('./components/accountPage/UserPanelLogic/UserPanelLogic'),
  );
  const DashboardLogic = lazy(
    () => import('./components/workspaces/DashboardLogic/DashboardLogic'),
  );

  return (
    <ThemeContextProvider>
      <Suspense fallback={<FullPageLoader />}>
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
          ]}
        />
      </Suspense>
    </ThemeContextProvider>
  );
}

export default App;
