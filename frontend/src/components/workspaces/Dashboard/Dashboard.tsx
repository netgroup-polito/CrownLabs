import { Col } from 'antd';
import { FC, useEffect, useState } from 'react';
import { Workspace } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceWelcome } from '../WorkspaceWelcome';

const dashboard = new SessionValue(StorageKeys.Dashboard_View, '-1');
export interface IDashboardProps {
  tenantNamespace: string;
  workspaces: Array<Workspace>;
}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWsId, setSelectedWs] = useState(parseInt(dashboard.get()));
  const { tenantNamespace, workspaces } = props;

  useEffect(() => {
    dashboard.set(String(selectedWsId));
  }, [selectedWsId]);

  return (
    <>
      <Col span={24} lg={8} xxl={8} className="lg:pr-4 py-5 lg:h-full flex">
        <div className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar">
          <WorkspaceGrid
            selectedWs={selectedWsId}
            workspaceItems={workspaces.map((ws, idx) => ({
              id: idx,
              title: ws.prettyName,
            }))}
            onClick={setSelectedWs}
          />
        </div>
      </Col>
      <Col
        span={24}
        lg={14}
        xxl={12}
        className="lg:pl-4 lg:pr-0 px-4 flex flex-auto"
      >
        {selectedWsId !== -1 ? (
          <WorkspaceContainer
            tenantNamespace={tenantNamespace}
            workspace={workspaces[selectedWsId]}
          />
        ) : (
          <WorkspaceWelcome />
        )}
      </Col>
    </>
  );
};

export default Dashboard;
