import { BarChartOutlined, UserOutlined } from '@ant-design/icons';
import { useContext, useCallback } from 'react';
import './App.css';
import { TenantContext } from './contexts/TenantContext';
import { QuotaProvider } from './contexts/QuotaContext';
import { LinkPosition } from './utils';
import AppLayout from './components/common/AppLayout';
import DashboardLogic from './components/workspaces/DashboardLogic/DashboardLogic';
import ActiveViewLogic from './components/activePage/ActiveViewLogic/ActiveViewLogic';
import UserPanelLogic from './components/accountPage/UserPanelLogic/UserPanelLogic';
import SSHTerminal from './components/activePage/SSHTerminal/SSHTerminal';
import { useOwnedInstancesQuery } from './generated-types';
import { useQuotaCalculations } from './components/workspaces/QuotaDisplay/useQuotaCalculation';
import { ErrorContext } from './errorHandling/ErrorContext';
import type { ApolloError } from '@apollo/client';

function App() {
  const { data: tenantData } = useContext(TenantContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const tenantNs = tenantData?.tenant?.status?.personalNamespace?.name;

  // Get all instances for quota calculations
  const { data: instancesData, refetch: refetchInstances } =
    useOwnedInstancesQuery({
      variables: { tenantNamespace: tenantNs || '' },
      skip: !tenantNs,
      onError: apolloErrorCatcher,
      fetchPolicy: 'cache-and-network',
    });

  // Calculate quota using our new hook
  // Filter out null instances before passing to the hook
  const validInstances = instancesData?.instanceList?.instances?.filter(
    (instance): instance is NonNullable<typeof instance> => instance != null,
  );

  // Handle null tenant by converting to undefined
  const tenant = tenantData?.tenant ?? undefined;

  const quotaCalculations = useQuotaCalculations(validInstances, tenant);

  // Enhanced refresh function
  const refreshQuota = useCallback(async () => {
    try {
      await refetchInstances();
    } catch (error) {
      console.error('Error refreshing quota data:', error);
      apolloErrorCatcher(error as ApolloError);
    }
  }, [refetchInstances, apolloErrorCatcher]);

  return (
    <QuotaProvider
      refreshQuota={refreshQuota}
      consumedQuota={quotaCalculations.consumedQuota}
      workspaceQuota={quotaCalculations.workspaceQuota}
      availableQuota={quotaCalculations.availableQuota}
    >
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
              path: '/instance/:namespace/:VMname/:environment/ssh',
            },
            content: (
              <SSHTerminal key="/instance/:namespace/:VMname/:environment/ssh" />
            ),
            linkPosition: LinkPosition.Hidden,
          },
        ]}
      />
    </QuotaProvider>
  );
}

export default App;
