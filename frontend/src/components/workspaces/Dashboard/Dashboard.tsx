import { FC, useState } from 'react';
import { Col } from 'antd';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceWelcome } from '../WorkspaceWelcome';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import { WorkspaceRole } from '../../../utils';
export interface IDashboardProps {
  tenantNamespace: string;
  workspaces: Array<{
    workspaceId: string;
    role: WorkspaceRole;
    workspaceNamespace: string;
  }>;
}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWsId, setSelectedWs] = useState(-1);
  const { tenantNamespace, workspaces } = props;
  const [reload, setReload] = useState(false);

  return (
    <>
      <Col span={24} lg={8} xxl={8} className="lg:pr-4 py-5 lg:h-full flex">
        <div className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar">
          <WorkspaceGrid
            selectedWs={selectedWsId}
            workspaceItems={workspaces.map((wk, idx) => ({
              id: idx,
              title: wk.workspaceId,
            }))}
            onClick={(id: number) => {
              setSelectedWs(id);
              setReload(true);
            }}
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
            workspace={{
              id: selectedWsId,
              role: workspaces[selectedWsId].role,
              title: workspaces[selectedWsId].workspaceId,
              workspaceNamespace: workspaces[selectedWsId].workspaceNamespace,
            }}
            reload={reload}
            setReload={setReload}
          />
        ) : (
          <WorkspaceWelcome />
        )}
      </Col>
    </>
  );
};

export default Dashboard;
