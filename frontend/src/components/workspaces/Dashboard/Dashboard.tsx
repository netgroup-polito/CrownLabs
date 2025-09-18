import { Col, Button } from 'antd';
import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import QuotaDisplay from '../QuotaDisplay/QuotaDisplay';
import WorkspaceAdd from '../WorkspaceAdd/WorkspaceAdd';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import { WorkspaceWelcome } from '../WorkspaceWelcome';
import { QuotaProvider } from '../../../contexts/QuotaContext';

const dashboard = new SessionValue(StorageKeys.Dashboard_View, '-1');
export interface IDashboardProps {
  tenantNamespace: string;
  tenantPersonalWorkspace?: {
    createPWs: boolean;
    isPWsCreated: boolean;
    quota: {
      cpu: string;
      memory: string;
      instances: number;
    } | null;
  };
  workspaces: Array<Workspace>;
  candidatesButton?: {
    show: boolean;
    selected: boolean;
    select: () => void;
  };
  globalQuota?: {
    consumedQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    workspaceQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    availableQuota: {
      cpu?: string | number;
      memory?: string;
      instances?: number;
    };
    showQuotaDisplay: boolean;
    refreshQuota?: () => void; // Add refresh function
  };
}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWsId, setSelectedWs] = useState(parseInt(dashboard.get()));
  const { tenantNamespace, workspaces, candidatesButton, globalQuota } = props;

  useEffect(() => {
    dashboard.set(String(selectedWsId));
  }, [selectedWsId]);

  // prepare IWorkspaceGridProps.workspaceItems
  const workspaceItems = useMemo(() => {
    return workspaces
      .map((ws, idx) => ({
        id: idx,
        title: ws.prettyName,
        waitingTenants: ws.waitingTenants,
      }))
      .sort((a, b) => a.title.localeCompare(b.title));
  }, [workspaces]);

  return (
    <QuotaProvider
      refreshQuota={globalQuota?.refreshQuota}
      availableQuota={globalQuota?.availableQuota}
    >
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          height: 'calc(100vh - 64px)', // Subtract navbar height
          gap: '16px',
          overflow: 'hidden', // Prevent overall container from growing
        }}
      >
        {/* Global Quota Display - Fixed Height */}
        {globalQuota?.showQuotaDisplay && globalQuota.workspaceQuota && (
          <div style={{ flexShrink: 0 }}>
            <QuotaDisplay
              consumedQuota={globalQuota.consumedQuota}
              workspaceQuota={globalQuota.workspaceQuota}
            />
          </div>
        )}

        {/* Dashboard Grid Layout - Scrollable */}
        <div
          style={{ display: 'flex', gap: '16px', flex: 1, overflow: 'hidden' }}
        >
          <Col
            span={24}
            lg={8}
            xxl={8}
            className="lg:pr-2 flex"
            style={{ height: '100%' }}
          >
            <div
              className="flex-auto overflow-auto scrollbar"
              style={{ height: '100%' }}
            >
              <WorkspaceGrid
                tenantPersonalWorkspace={props.tenantPersonalWorkspace}
                selectedWs={selectedWsId}
                workspaceItems={workspaceItems}
                onClick={setSelectedWs}
              />
              {candidatesButton?.show && (
                <div className="lg:mt-4 mt-0 text-center">
                  <Button
                    shape="round"
                    size={'middle'}
                    onClick={candidatesButton.select}
                  >
                    {candidatesButton.selected ? 'Hide' : 'Load'} candidates
                  </Button>
                </div>
              )}
            </div>
          </Col>
          <Col
            span={24}
            lg={14}
            xxl={12}
            className="lg:pl-4 lg:pr-0 px-4 flex"
            style={{ height: '100%', overflow: 'hidden' }}
          >
            {/* Workspace Container content */}
            {selectedWsId >= 0 && selectedWsId < workspaces.length ? (
              <WorkspaceContainer
                tenantNamespace={tenantNamespace}
                workspace={workspaces[selectedWsId]}
                availableQuota={globalQuota?.availableQuota}
                refreshQuota={globalQuota?.refreshQuota}
                isPersonalWorkspace={false}
              />
            ) : selectedWsId === -1 ? (
              <WorkspaceContainer
                tenantNamespace={tenantNamespace}
                workspace={{
                  name: 'personal-frontend-only',
                  prettyName: 'Personal Workspace',
                  role: WorkspaceRole.manager,
                  namespace: tenantNamespace,
                  waitingTenants: undefined,
                }}
                availableQuota={globalQuota?.availableQuota}
                refreshQuota={globalQuota?.refreshQuota}
                isPersonalWorkspace={true}
              />
            ) : selectedWsId === -2 ? (
              <WorkspaceAdd />
            ) : (
              <WorkspaceWelcome />
            )}
          </Col>
        </div>
      </div>
    </QuotaProvider>
  );
};

export default Dashboard;
