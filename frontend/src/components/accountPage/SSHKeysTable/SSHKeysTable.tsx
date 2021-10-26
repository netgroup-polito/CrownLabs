import { DeleteOutlined } from '@ant-design/icons';
import { Popconfirm, Table } from 'antd';
import Column from 'antd/lib/table/Column';
import { FC } from 'react';

export interface ISSHKeysTableProps {
  sshKeys?: { name: string; key: string }[];
  onDeleteKey?: (key: { name: string; key: string }) => void;
}

const SSHKeysTable: FC<ISSHKeysTableProps> = props => {
  const { sshKeys } = props;
  return (
    <Table
      dataSource={sshKeys}
      expandedRowRender={record => <p>{record.key}</p>}
      style={{ maxWidth: '800px' }}
    >
      <Column title="Name" dataIndex="name" width={100} />
      <Column title="Key" dataIndex="key" ellipsis={true} width={240} />
      <Column
        title="Action"
        key="x"
        width={60}
        render={(_: any, record: { name: string; key: string }) =>
          sshKeys?.length && (
            <Popconfirm
              className="flex justify-center"
              title="Confirm deletion?"
              onConfirm={() => props.onDeleteKey?.(record)}
            >
              <DeleteOutlined style={{ color: 'red' }} />
            </Popconfirm>
          )
        }
      />
    </Table>
  );
};

export default SSHKeysTable;
