import './App.css';
import { Alert, Skeleton } from 'antd';
import Box from './components/common/Box';
import AppLayout from './components/common/AppLayout';
import ThemeContextProvider from './contexts/ThemeContext';
import { BarChartOutlined } from '@ant-design/icons';
import { AuthContext } from './contexts/AuthContext';
import { useContext } from 'react';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
function App() {
  const { userId } = useContext(AuthContext);
  return (
    <ThemeContextProvider>
      <AppLayout
        TooltipButtonLink={
          'https://grafana.crownlabs.polito.it/d/BOZGskUGz/personal-overview?&var-user=' +
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
          ...[
            { name: 'Active', path: '/active' },
            {
              name: 'Drive',
              path: 'https://crownlabs.polito.it/cloud/apps/dashboard/',
              externalLink: true,
            },
          ].map(r => {
            return {
              route: {
                ...r,
              },
              content: (
                <Box
                  header={{
                    size: 'middle',
                    center: (
                      <div className="h-full flex justify-center items-center px-5">
                        <p className="md:text-2xl text-xl text-center mb-0">
                          <b>{r.name}</b>
                        </p>
                      </div>
                    ),
                  }}
                >
                  <div className="flex justify-center">
                    <Alert
                      className="mb-4 mt-8 mx-8 w-full"
                      message="Warning"
                      description="This is a temporary content"
                      type="warning"
                      showIcon
                      closable
                    />
                  </div>
                  <Skeleton className="px-8 pt-1" />
                </Box>
              ),
            };
          }),
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
