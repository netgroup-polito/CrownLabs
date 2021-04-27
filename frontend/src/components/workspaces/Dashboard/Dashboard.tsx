import { FC, useState } from 'react';
import { Row, Col } from 'antd';
import './Dashboard.css';
import { Workspace, WorkspaceWelcome } from '../Workspace';
import Courses from '../Courses/Courses';
import data from '../FakeData';
import '../../../index.less'; //To delete, usefull only to storybook

export interface IDashboardProps {}

const Dashboard: FC<IDashboardProps> = () => {
  const [selected, setSelected] = useState(-1);
  const courses = data.map(workspace =>
    Object.assign({}, { id: workspace.id, title: workspace.title })
  );
  const workspace = data.find(workspace => workspace.id === selected);

  return (
    <>
      <Row className="h-full py-10 flex">
        <Col span={0} lg={1} xxl={2}></Col>
        <Col span={24} lg={8} xxl={8} className="pr-4 px-4 py-5 lg:h-full flex">
          <div className="flex-auto lg:overflow-x-hidden overflow-auto scrollbar">
            <Courses
              selected={selected}
              courses={courses}
              onClick={setSelected}
            />
          </div>
        </Col>
        <Col span={24} lg={14} xxl={12} className="px-4 flex flex-auto">
          {workspace ? (
            <Workspace workspace={workspace} />
          ) : (
            <WorkspaceWelcome />
          )}
        </Col>
        <Col span={0} lg={1} xxl={2}></Col>
      </Row>
    </>
  );
};

export default Dashboard;
