import { FC, useState } from 'react';
import { Table, Modal } from 'antd';
import Button from 'antd-button-color';
import InstanceIcons, {
  IInstanceIconsProps,
} from '../InstanceIcons/InstanceIcons';
import InstanceContent from '../InstanceContent/InstanceContent';
import SSHModalContent from './SSHModalContent/SSHModalContent';
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

export interface ISSHInfo {
  username: string;
  password: string;
  ip: string;
}

const InstancesTable: FC<IInstancesTableProps> = ({ ...props }) => {
  const { instances, isManaged } = props;
  const { Column } = Table;
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [sshInfo, setSshInfo] = useState({
    username: '',
    password: '',
    ip: '',
  } as ISSHInfo);
  const data = instances.map((obj: IInstance) => {
    return { key: obj.id, icons: 'VM icons', ...obj };
  });
  return (
    <>
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
            <InstanceIcons isGUI={record.isGUI} phase={record.phase} />
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
              ip={record.ip}
              phase={record.phase}
              toggleModal={setIsModalOpen}
              setSshInfo={setSshInfo}
            />
          )}
        />
      </Table>
      <Modal
        title="SSH Connection"
        visible={isModalOpen}
        onOk={() => setIsModalOpen(false)}
        onCancel={() => setIsModalOpen(false)}
        footer={[<Button onClick={() => setIsModalOpen(false)}>Close</Button>]}
        centered
      >
        <SSHModalContent sshInfo={sshInfo} />
      </Modal>
    </>
  );
};

export default InstancesTable;
