import { FC, useState } from 'react';
import { Table } from 'antd';
import { Instance, WorkspaceRole } from '../../../utils';
import { FetchResult } from '@apollo/client';
import { DeleteInstanceMutation } from '../../../generated-types';
import './TableInstance.less';
import RowInstanceTitle from './RowInstanceTitle/RowInstanceTitle';
import RowInstanceActions from './RowInstanceActions/RowInstanceActions';

const { Column } = Table;
export interface ITableInstanceProps {
  viewMode: WorkspaceRole;
  instances: Array<Instance>;
  showGuiIcon: boolean;
  extended: boolean;
  startInstance?: (idInstance: string, idTemplate: string) => void;
  stopInstance?: (idInstance: string, idTemplate: string) => void;
  destroyInstance?: (
    tenantNamespace: string,
    instanceId: string
  ) => Promise<
    FetchResult<
      DeleteInstanceMutation,
      Record<string, any>,
      Record<string, any>
    >
  >;
}

const TableInstance: FC<ITableInstanceProps> = ({ ...props }) => {
  const {
    instances,
    viewMode,
    extended,
    showGuiIcon,
    startInstance,
    stopInstance,
    destroyInstance,
  } = props;

  const [now, setNow] = useState(new Date());

  setInterval(() => setNow(new Date()), 60000);

  const data = instances;
  return (
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
        rowKey={record =>
          extended && viewMode === 'user'
            ? record.id + record.idTemplate!
            : record.id
        }
      >
        <Column
          className={
            extended
              ? viewMode === WorkspaceRole.user
                ? 'w-5/6 sm:w-2/3 lg:w-3/5 xl:w-1/2 2xl:w-5/12'
                : 'w-3/5 md:w-1/2'
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
                ? 'w-1/6 sm:w-1/3 lg:w-2/5 xl:w-1/22xl:w-7/12'
                : 'w-2/5 md:w-1/2'
              : 'w-1/3 md:w-1/4'
          }
          title="Instance Actions"
          key="actions"
          render={(instance: Instance) => (
            <RowInstanceActions
              instance={instance}
              ssh={{ IP: instance.ip, KEY: '' }}
              now={now}
              fileManager={true}
              extended={extended}
              startInstance={startInstance}
              stopInstance={stopInstance}
              destroyInstance={() =>
                destroyInstance!(instance.tenantNamespace!, instance.name)
              }
            />
          )}
        />
      </Table>
    </div>
  );
};

export default TableInstance;
