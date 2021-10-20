import { FC, useState } from 'react';
import { Table } from 'antd';
import { Instance, WorkspaceRole } from '../../../utils';
import './TableInstance.less';
import RowInstanceTitle from './RowInstanceTitle/RowInstanceTitle';
import RowInstanceActions from './RowInstanceActions/RowInstanceActions';

const { Column } = Table;
export interface ITableInstanceProps {
  viewMode: WorkspaceRole;
  instances: Array<Instance>;
  showGuiIcon: boolean;
  extended: boolean;
}

const TableInstance: FC<ITableInstanceProps> = ({ ...props }) => {
  const { instances, viewMode, extended, showGuiIcon } = props;

  const [now, setNow] = useState(new Date());

  setInterval(() => setNow(new Date()), 60000);

  const startInstance = (idInstance: number, idTemplate: string) => {};
  const stopInstance = (idInstance: number, idTemplate: string) => {};
  const destroyInstance = (idInstance: number, idTemplate: string) => {};

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
              ? viewMode === 'user'
                ? 'w-2/3 md:w-1/2 lg:w-5/12'
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
              ? viewMode === 'user'
                ? 'w-1/3 md:w-1/2 lg:w-7/12'
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
              destroyInstance={destroyInstance}
            />
          )}
        />
      </Table>
    </div>
  );
};

export default TableInstance;
