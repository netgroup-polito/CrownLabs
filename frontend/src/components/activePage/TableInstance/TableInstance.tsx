import { FC, useState } from 'react';
import { Popconfirm, Table } from 'antd';
import { Instance, WorkspaceRole } from '../../../utils';
import { useDeleteInstanceMutation } from '../../../generated-types';
import './TableInstance.less';
import RowInstanceTitle from './RowInstanceTitle/RowInstanceTitle';
import RowInstanceActions from './RowInstanceActions/RowInstanceActions';
import Button from 'antd-button-color';
import { DeleteOutlined } from '@ant-design/icons';

const { Column } = Table;
export interface ITableInstanceProps {
  viewMode: WorkspaceRole;
  instances: Array<Instance>;
  hasSSHKeys?: boolean;
  showGuiIcon: boolean;
  extended: boolean;
}

const TableInstance: FC<ITableInstanceProps> = ({ ...props }) => {
  const { instances, viewMode, extended, hasSSHKeys, showGuiIcon } = props;

  const [now, setNow] = useState(new Date());
  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  setInterval(() => setNow(new Date()), 60000);

  const destroyAll = () => {
    instances.forEach(instance => {
      deleteInstanceMutation({
        variables: {
          instanceId: instance.name,
          tenantNamespace: instance.tenantNamespace!,
        },
      });
    });
  };

  const data = instances;
  return (
    <>
      <div
        className={`rowInstance-bg-color ${
          viewMode === 'user' && extended
            ? 'cl-table-instance flex-grow flex-wrap content-between py-0 overflow-auto scrollbar'
            : ''
        }`}
      >
        <Table
          className="rowInstance-bg-color"
          dataSource={data}
          showHeader={false}
          pagination={false}
          size={'middle'}
          rowClassName={
            viewMode === 'user' && extended ? '' : 'rowInstance-bg-color'
          }
          rowKey={record => record.id + record.idTemplate!}
        >
          <Column
            className={
              extended
                ? viewMode === WorkspaceRole.user
                  ? 'w-5/6 sm:w-2/3 lg:w-3/5 xl:w-1/2 2xl:w-5/12'
                  : 'w-3/5 sm:w-2/3 lg:w-1/2 xl:w-1/2'
                : 'w-2/3 md:w-3/4'
            }
            title="Instance Title"
            key="title"
            render={(instance: Instance) => (
              <RowInstanceTitle
                viewMode={viewMode}
                extended={extended}
                instance={instance}
                showGuiIcon={showGuiIcon}
              />
            )}
          />
          <Column
            className={
              extended
                ? viewMode === WorkspaceRole.user
                  ? 'w-1/6 sm:w-1/3 lg:w-2/5 xl:w-1/2 2xl:w-7/12'
                  : 'w-2/5 sm:w-1/3 lg:w-1/2 xl:w-1/2'
                : 'w-1/3 md:w-1/4'
            }
            title="Instance Actions"
            key="actions"
            render={(instance: Instance) => (
              <RowInstanceActions
                instance={instance}
                hasSSHKeys={hasSSHKeys}
                now={now}
                fileManager={true}
                extended={extended}
                viewMode={viewMode}
              />
            )}
          />
        </Table>
      </div>
      {extended && viewMode === WorkspaceRole.user && (
        <div className="w-full pt-5 flex justify-center ">
          <Popconfirm
            placement="left"
            title="You are about to delete all VMs in this. Are you sure?"
            okText="Yes"
            cancelText="No"
            onConfirm={destroyAll}
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
      )}
    </>
  );
};

export default TableInstance;
