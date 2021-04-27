import { FC } from 'react';
import { Table } from 'antd';

import InstancesTableRow from './InstancesTableRow';

export interface IInstancesTableProps {
  id: string;
  instances: Array<{ id: number; name: string; ip: string; status: boolean }>;
  destroyInstance: (idInstance: number, idTemplate: string) => void;
}

const InstancesTable: FC<IInstancesTableProps> = ({ ...props }) => {
  const { id, instances, destroyInstance } = props;
  const instancesRows = instances.map(instance =>
    Object.assign(
      {},
      {
        key: instance.id,
        instance: (
          <InstancesTableRow
            idInstance={instance.id}
            idTemplate={id}
            name={instance.name}
            ip={instance.ip}
            status={instance.status}
            destroyInstance={destroyInstance}
          />
        ),
      }
    )
  );

  const columns = [
    {
      title: 'Instance',
      dataIndex: 'instance',
      key: 'instance',
    },
  ];

  return (
    <Table
      size={'small'}
      showHeader={false}
      dataSource={instancesRows}
      pagination={false}
      columns={columns}
    />
  );
};

export default InstancesTable;
