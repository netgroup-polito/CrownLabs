import { Button, Col } from 'antd';
import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceWelcome } from '../WorkspaceWelcome';
import WorkspaceAdd from '../WorkspaceAdd/WorkspaceAdd';
import QuotaDisplay from '../QuotaDisplay/QuotaDisplay';

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
  };
}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWsId, setSelectedWs] = useState(parseInt(dashboard.get()));
  const { tenantNamespace, workspaces, candidatesButton, globalQuota } = props;

  useEffect(() => {
    dashboard.set(String(selectedWsId));
  }, [selectedWsId]);

  console.log('Dashboard tenantNamespace:', tenantNamespace);
  console.log('Dashboard selectedWsId:', selectedWsId);
  console.log('Dashboard workspaces:', workspaces);

  // Track when tenantNamespace changes
  useEffect(() => {
    console.log(
      'Dashboard tenantNamespace changed to:',
      tenantNamespace,
      'Stack:',
      new Error().stack,
    );
  }, [tenantNamespace]);

  // Track when workspaces array changes
  useEffect(() => {
    console.log(
      'Dashboard workspaces array changed:',
      workspaces,
      'Stack:',
      new Error().stack,
    );
  }, [workspaces]);

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
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        gap: '5vh',
      }}
    >
      {/* Global Quota Display - Full Width */}
      {globalQuota?.showQuotaDisplay && globalQuota.workspaceQuota && (
        <QuotaDisplay
          consumedQuota={globalQuota.consumedQuota}
          workspaceQuota={globalQuota.workspaceQuota}
        />
      )}

      {/* Dashboard Grid Layout */}
      <div style={{ display: 'flex', gap: '16px' }}>
        <Col
          span={24}
          lg={8}
          xxl={8}
          className="lg:pr-2 lg:pt-2 lg:pb-0 py-5 lg:h-full flex"
        >
          <div className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar lg:h-full">
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
          className="lg:pl-4 lg:pr-0 px-4 flex flex-auto"
        >
          {selectedWsId >= 0 && selectedWsId < workspaces.length ? (
            <>
              {(() => {
                console.log(
                  'Dashboard rendering WorkspaceContainer for regular workspace:',
                  {
                    selectedWsId,
                    workspace: workspaces[selectedWsId],
                    tenantNamespace,
                    workspaceNamespace: workspaces[selectedWsId].namespace,
                    isPersonal: false,
                  },
                );
                return null;
              })()}
              <WorkspaceContainer
                tenantNamespace={tenantNamespace}
                workspace={workspaces[selectedWsId]}
                availableQuota={globalQuota?.availableQuota}
                isPersonalWorkspace={false}
              />
            </>
          ) : selectedWsId === -1 ? (
            <>
              {(() => {
                // Create a completely isolated personal workspace object
                const personalWorkspace = {
                  name: 'personal-frontend-only', // Different name to avoid Apollo conflicts
                  prettyName: 'Personal Workspace',
                  role: WorkspaceRole.manager,
                  namespace: tenantNamespace, // Always use tenantNamespace directly
                  waitingTenants: undefined,
                };
                console.log(
                  'Dashboard rendering WorkspaceContainer for personal workspace:',
                  {
                    selectedWsId,
                    tenantNamespace,
                    workspaceNamespace: tenantNamespace, // Always use tenantNamespace
                    isPersonal: true,
                    workspace: personalWorkspace,
                  },
                );
                return null;
              })()}
              <WorkspaceContainer
                tenantNamespace={tenantNamespace}
                workspace={{
                  name: 'personal-frontend-only', // Different name to avoid Apollo conflicts
                  prettyName: 'Personal Workspace',
                  role: WorkspaceRole.manager,
                  namespace: tenantNamespace, // Always use tenantNamespace directly
                  waitingTenants: undefined,
                }}
                availableQuota={globalQuota?.availableQuota}
                isPersonalWorkspace={true}
              />
            </>
          ) : selectedWsId === -2 ? (
            <WorkspaceAdd />
          ) : (
            <WorkspaceWelcome />
          )}
        </Col>
      </div>
    </div>
  );
};

export default Dashboard;
