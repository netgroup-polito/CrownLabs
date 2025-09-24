import { Col, Button, Row } from 'antd';
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
      <div className="flex flex-col lg:h-full h-auto min-h-0">
        <div
          className="w-full px-4 flex flex-col min-h-0"
          style={{ height: '100%', minHeight: 0 }}
        >
          {globalQuota?.showQuotaDisplay && globalQuota.workspaceQuota && (
            <div style={{ flexShrink: 0 }}>
              <QuotaDisplay
                consumedQuota={globalQuota.consumedQuota}
                workspaceQuota={globalQuota.workspaceQuota}
              />
            </div>
          )}

          <Row
            gutter={16}
            className="flex-1 lg:h-full min-h-0"
            align="stretch"
            style={{ minHeight: 0 }}
          >
            <Col
              span={24}
              lg={8}
              xxl={8}
              className="lg:pr-2 lg:pt-2 lg:pb-0 py-5 lg:h-full flex"
            >
              <div
                className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar lg:h-full"
                style={{ minHeight: 0 }}
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
              lg={16} // <-- make columns sum to 24 (8 + 16)
              xxl={16}
              className="lg:pl-4 lg:pr-0 px-4 flex flex-auto"
            >
              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  flex: '1 1 auto',
                  minHeight: 0,
                }}
              >
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
              </div>
            </Col>
          </Row>
        </div>
      </div>
    </QuotaProvider>
  );
};

export default Dashboard;
