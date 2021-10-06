import { Table } from 'antd';
import Column from 'antd/lib/table/Column';
import { FC } from 'react';

export interface ISSHKeysTableProps {
  sshKeys?: { name: string; key: string }[];
}

const SSHKeysTable: FC<ISSHKeysTableProps> = props => {
  const { sshKeys } = props;
  return (
    <Table
      dataSource={sshKeys}
      expandedRowRender={record => <p>{record.key}</p>}
      style={{ maxWidth: '800px' }}
    >
      <Column title="Name" dataIndex="name" width={120} />
      <Column title="Key" dataIndex="key" ellipsis={true} />
    </Table>
  );
};

export default SSHKeysTable;
