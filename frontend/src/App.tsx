import './App.css';
import AppLayout from './components/common/AppLayout';
import ThemeContextProvider from './contexts/ThemeContext';
import { BarChartOutlined } from '@ant-design/icons';
import { AuthContext } from './contexts/AuthContext';
import { useContext } from 'react';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';

function App() {
  const { userId } = useContext(AuthContext);

  return (
    <ThemeContextProvider>
      <AppLayout
        TooltipButtonLink={
          'https://grafana.crownlabs.polito.it/d/BOZGskUGz/personal-overview?&var-namespace=' +
          userId
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
