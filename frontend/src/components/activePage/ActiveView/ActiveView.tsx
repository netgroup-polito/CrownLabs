import { FC, useState } from 'react';
import { Space, Col, Input, Popconfirm } from 'antd';
import ViewModeButton from './ViewModeButton/ViewModeButton';
import Box from '../../common/Box';
import { WorkspaceRole } from '../../../utils';
import { DeleteOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import TableInstanceLogic from '../TableInstance/TableInstanceLogic';
import TableWorkspaceLogic from '../TableWorkspaceLogic/TableWorkspaceLogic';
export interface IActiveViewProps {
  userId: string;
  tenantNamespace: string;
  workspaces: Array<{
    prettyName: string;
    role: WorkspaceRole;
    namespace: string;
    id: string;
  }>;
  managerView: boolean;
}

const ActiveView: FC<IActiveViewProps> = ({ ...props }) => {
  const { managerView, userId, tenantNamespace, workspaces } = props;
  const [searchField, setSearchField] = useState('');
  const [currentView, setCurrentView] = useState<WorkspaceRole>(
    WorkspaceRole.user
  );

  const { Search } = Input;
  return (
    <Col span={24} lg={22} xxl={20}>
      <Box
        header={{
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
          left: currentView === 'manager' && (
            <div className="h-full flex justify-center items-center pl-10">
              <Search
                className="hidden sm:block"
                placeholder="Filter by user ID"
                onSearch={value => {
                  setSearchField(value);
                }}
                enterButton
              />
            </div>
          ),
        }}
        footer={
          currentView === 'user' && (
            <div className="w-full py-5 flex justify-center ">
              <Popconfirm
                placement="left"
                title="You are about to delete all VMs in this. Are you sure?"
                okText="Yes"
                cancelText="No"
                onConfirm={e => e?.stopPropagation()}
                onCancel={e => e?.stopPropagation()}
              >
                <Button
                  type="danger"
                  shape="round"
                  size="large"
                  icon={<DeleteOutlined />}
                  onClick={e => e.stopPropagation()}
                >
                  Destory All
                </Button>
              </Popconfirm>
            </div>
          )
        }
      >
        {currentView === 'manager' && managerView ? (
          <TableWorkspaceLogic
            workspaces={workspaces}
            userId={userId}
            tenantNamespace={tenantNamespace}
            filter={searchField}
          />
        ) : (
          <TableInstanceLogic
            showGuiIcon={true}
            viewMode={currentView}
            extended={true}
            tenantNamespace={tenantNamespace}
          />
        )}
      </Box>
    </Col>
  );
};

export default ActiveView;
