import './App.css';
import AppLayout from './components/common/AppLayout';
import ThemeContextProvider from './contexts/ThemeContext';
import { BarChartOutlined } from '@ant-design/icons';
import { useContext } from 'react';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import { TenantContext } from './graphql-components/tenantContext/TenantContext';

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
          },
          {
            route: { name: 'Active', path: '/active' },
            content: <ActiveViewLogic />,
          },
          {
            route: { name: 'Drive', path: 'https://crownlabs.polito.it/cloud' },
          },
          {
            route: { name: 'Account', path: '/account' },
            content: <UserPanelLogic />,
          },
        ]}
      />
    </ThemeContextProvider>
  );
}

export default App;
