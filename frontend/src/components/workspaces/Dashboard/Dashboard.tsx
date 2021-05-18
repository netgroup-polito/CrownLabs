import { FC, useState } from 'react';
import { Row, Col } from 'antd';
import { WorkspaceContainer } from '../WorkspaceContainer';
import { WorkspaceWelcome } from '../WorkspaceWelcome';
import { WorkspaceGrid } from '../Grid/WorkspaceGrid';
import data from '../FakeData';
export interface IDashboardProps {}

const Dashboard: FC<IDashboardProps> = ({ ...props }) => {
  const [selectedWs, setSelectedWs] = useState(-1);
  const workspaceItems = data.map(workspace =>
    Object.assign({}, { id: workspace.id, title: workspace.title })
  );
  const workspace = data.find(workspace => workspace.id === selectedWs);

  return (
    <Row className="h-full py-10 flex">
      <Col span={0} lg={1} xxl={2}></Col>
      <Col span={24} lg={8} xxl={8} className="pr-4 px-4 py-5 lg:h-full flex">
        <div className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar">
          <WorkspaceGrid
            selectedWs={selectedWs}
            workspaceItems={workspaceItems}
            onClick={setSelectedWs}
          />
        </div>
      </Col>
      <Col span={24} lg={14} xxl={12} className="px-4 flex flex-auto">
        {workspace ? (
          <WorkspaceContainer workspace={workspace} />
        ) : (
          <WorkspaceWelcome />
        )}
      </Col>
      <Col span={0} lg={1} xxl={2}></Col>
    </Row>
  );
};

export default Dashboard;
