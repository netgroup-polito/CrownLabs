import { Col, Button } from 'antd';
import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import WorkspaceAdd from '../WorkspaceAdd/WorkspaceAdd';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import { WorkspaceWelcome } from '../WorkspaceWelcome';

const dashboard = new SessionValue(StorageKeys.Dashboard_View, '-3');
export interface IDashboardProps {
  tenantNamespace: string;
  tenantPersonalWorkspace?: {
    createPWs: boolean;
    isPWsCreated: boolean;
  };
  workspaces: Array<Workspace>;
  candidatesButton?: {
    show: boolean;
    selected: boolean;
    select: () => void;
  };
}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWsId, setSelectedWs] = useState(parseInt(dashboard.get()));
  const { tenantNamespace, workspaces, candidatesButton } = props;

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
    <>
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
        className="lg:pl-4 lg:pr-0 px-4 flex flex-auto h-full"
      >
        {selectedWsId >= 0 && selectedWsId < workspaces.length ? (
          <WorkspaceContainer
            tenantNamespace={tenantNamespace}
            workspace={workspaces[selectedWsId]}
            isPersonalWorkspace={false}
          />
        ) : selectedWsId === -1 ? (
          <WorkspaceContainer
            tenantNamespace={tenantNamespace}
            workspace={{
              name: 'personal',
              prettyName: 'Personal Workspace',
              role: WorkspaceRole.manager,
              namespace: tenantNamespace,
              waitingTenants: undefined,
            }}
            isPersonalWorkspace={true}
          />
        ) : selectedWsId === -2 ? (
          <WorkspaceAdd />
        ) : (
          <WorkspaceWelcome />
        )}
      </Col>
    </>
  );
};

export default Dashboard;
