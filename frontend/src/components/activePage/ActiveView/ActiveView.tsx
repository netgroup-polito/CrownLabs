import { FC, useEffect, useState } from 'react';
import { Space, Col, Input, Tooltip } from 'antd';
import ViewModeButton from './ViewModeButton/ViewModeButton';
import Box from '../../common/Box';
import { User, WorkspaceRole } from '../../../utils';
import TableInstanceLogic from '../TableInstance/TableInstanceLogic';
import TableWorkspaceLogic from '../TableWorkspaceLogic/TableWorkspaceLogic';
import Button from 'antd-button-color';
import { FullscreenExitOutlined, FullscreenOutlined } from '@ant-design/icons';

const { Search } = Input;
export interface IActiveViewProps {
  user: User;
  workspaces: Array<{
    prettyName: string;
    role: WorkspaceRole;
    namespace: string;
    id: string;
  }>;
  managerView: boolean;
}

const ActiveView: FC<IActiveViewProps> = ({ ...props }) => {
  const { managerView, user, workspaces } = props;
  const [expandAll, setExpandAll] = useState(false);
  const [collapseAll, setCollapseAll] = useState(false);
  const [searchField, setSearchField] = useState('');
  const [currentView, setCurrentView] = useState<WorkspaceRole>(
    (window.sessionStorage.getItem('prevViewActivePage') as WorkspaceRole) ??
      WorkspaceRole.user
  );

  useEffect(() => {
    window.sessionStorage.setItem('prevViewActivePage', currentView);
  }, [currentView]);

  return (
    <Col span={24} lg={22} xxl={20}>
      <Box
        header={{
          center: !managerView && (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Active Instances</b>
              </p>
            </div>
          ),
          size: 'middle',
          right: managerView && (
            <div className="h-full flex justify-center items-center pr-10">
              <Space size="small">
                <ViewModeButton
                  setCurrentView={setCurrentView}
                  currentView={currentView}
                />
              </Space>
            </div>
          ),
          left: currentView === WorkspaceRole.manager && (
            <div className="h-full flex justify-center items-center pl-10 gap-4">
              <Search
                className="hidden sm:block"
                placeholder="Search User"
                onChange={event => {
                  setSearchField(event.target.value);
                }}
                onSearch={value => {
                  setSearchField(value);
                }}
                enterButton
              />
              <Button
                className="hidden xl:block"
                type="primary"
                shape="round"
                size="middle"
                icon={<FullscreenOutlined />}
                onClick={() => {
                  setExpandAll(true);
                }}
              >
                Expand
              </Button>
              <Button
                className="hidden xl:block"
                type="dark"
                shape="round"
                size="middle"
                icon={<FullscreenExitOutlined />}
                onClick={() => {
                  setCollapseAll(true);
                }}
              >
                Collapse
              </Button>
              <Tooltip title="Expand All" placement="top">
                <Button
                  className="xl:hidden"
                  type="primary"
                  shape="circle"
                  size="middle"
                  icon={<FullscreenOutlined />}
                  onClick={() => {
                    setExpandAll(true);
                  }}
                />
              </Tooltip>
              <Tooltip title="Collapse All" placement="top">
                <Button
                  className="xl:hidden"
                  type="dark"
                  shape="circle"
                  size="middle"
                  icon={<FullscreenExitOutlined />}
                  onClick={() => {
                    setCollapseAll(true);
                  }}
                />
              </Tooltip>
            </div>
          ),
        }}
      >
        {currentView === WorkspaceRole.manager && managerView ? (
          <TableWorkspaceLogic
            workspaces={workspaces}
            user={user}
            filter={searchField}
            collapseAll={collapseAll}
            expandAll={expandAll}
            setCollapseAll={setCollapseAll}
            setExpandAll={setExpandAll}
          />
        ) : (
          <TableInstanceLogic
            showGuiIcon={true}
            user={user}
            viewMode={currentView}
            extended={true}
          />
        )}
      </Box>
    </Col>
  );
};

export default ActiveView;
