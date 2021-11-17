import { FC, useState } from 'react';
import { Space, Col, Input } from 'antd';
import ViewModeButton from './ViewModeButton/ViewModeButton';
import Box from '../../common/Box';
import { User, WorkspaceRole } from '../../../utils';
import TableInstanceLogic from '../TableInstance/TableInstanceLogic';
import TableWorkspaceLogic from '../TableWorkspaceLogic/TableWorkspaceLogic';

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
  const [searchField, setSearchField] = useState('');
  const [currentView, setCurrentView] = useState<WorkspaceRole>(
    WorkspaceRole.user
  );

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
            <div className="h-full flex justify-center items-center pl-10">
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
            </div>
          ),
        }}
      >
        {currentView === WorkspaceRole.manager && managerView ? (
          <TableWorkspaceLogic
            workspaces={workspaces}
            user={user}
            filter={searchField}
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
