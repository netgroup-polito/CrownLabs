import { FC, useState } from 'react';
import { Space, Col, Input, Popconfirm } from 'antd';
import TableInstance from '../TableInstance/TableInstance';
import TableWorkspace from '../TableWorkspace/TableWorkspace';
import ViewModeButton from './ViewModeButton/ViewModeButton';
import Box from '../../common/Box';
import { WorkspaceRole, Workspace, Instance } from '../../../utils';
import { DeleteOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
export interface IActiveViewProps {
  workspaces: Array<Workspace> | [];
  instances: Array<Instance> | [];
  managerView: boolean;
}

const ActiveView: FC<IActiveViewProps> = ({ ...props }) => {
  const { managerView, workspaces, instances } = props;
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
          <TableWorkspace
            workspaces={workspaces.filter(
              ({ role }) => role === WorkspaceRole.manager
            )}
            filter={searchField}
          />
        ) : (
          <TableInstance
            showGuiIcon={true}
            viewMode={currentView}
            instances={instances}
            extended={true}
          />
        )}
      </Box>
    </Col>
  );
};

export default ActiveView;
