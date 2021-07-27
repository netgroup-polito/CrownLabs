import { FC } from 'react';
import { Table } from 'antd';
import VmInstanceIcons, {
  IInstanceIconsProps,
} from '../InstanceIcons/InstanceIcons';
import InstanceContent from '../InstanceContent/InstanceContent';
import './InstancesTable.css';

export interface IInstance {
  id: string;
  templateId: string;
  tenantId: string;
  tenantDisplayName: string;
  displayName: string;
  phase: 'ready' | 'creating' | 'failed' | 'stopping' | 'off';
  ip: string;
  cliOnly: boolean;
}

export interface IInstancesTableProps {
  isManaged: boolean;
  instances: Array<IInstance>;
}

const InstancesTable: FC<IInstancesTableProps> = ({ ...props }) => {
  const { instances, isManaged } = props;
  const { Column } = Table;
  const data = instances.map((obj: IInstance) => {
    return { key: obj.id, icons: 'VM icons', ...obj };
  });
  return (
    <Table
      dataSource={data}
      showHeader={false}
      pagination={false}
      size="small"
      rowClassName="rowInstance-bg-color"
    >
      <Column
        title="Icons"
        dataIndex="icons"
        key="icons"
        align="left"
        width={100}
        className="py-2"
        responsive={['md']}
        render={(text, record: IInstanceIconsProps) => (
          <VmInstanceIcons isGUI={record.isGUI} phase={record.phase} />
        )}
      />
      <Column
        title="Title"
        dataIndex="title"
        key="title"
        className="py-2"
        render={(text, record: IInstance) => (
          <InstanceContent
            displayName={record.displayName}
            tenantId={record.tenantId}
            tenantDisplayName={record.tenantDisplayName}
            isManaged={isManaged}
          />
        )}
      />
    </Table>
  );
};

export default InstancesTable;
