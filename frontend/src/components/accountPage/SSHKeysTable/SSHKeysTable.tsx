import { DeleteOutlined } from '@ant-design/icons';
import { Button, Table } from 'antd';
import { type FC, useState } from 'react';
import { ModalAlert } from '../../common/ModalAlert';

const Column = Table.Column;

export interface ISSHKeysTableProps {
  sshKeys?: { name: string; key: string }[];
  onDeleteKey: (key: { name: string; key: string }) => Promise<boolean>;
}

const SSHKeysTable: FC<ISSHKeysTableProps> = props => {
  const { sshKeys, onDeleteKey } = props;
  const [recordToDelete, setRecordToDelete] = useState<{
    name: string;
    key: string;
  } | null>(null);
  return (
    <Table
      dataSource={sshKeys}
      expandedRowRender={record => <p>{record.key}</p>}
      style={{ maxWidth: '800px' }}
      locale={{
        emptyText: (
          <div>
            <div>It seems that you don't have any SSH key registered</div>
            <div>
              If you don't know how to generate and upload a new key follow{` `}
              <a
                target="_blank"
                rel="noreferrer"
                href="https://crownlabs.polito.it/resources/crownlabs_ssh/"
              >
                this guide
              </a>
              .
            </div>
          </div>
        ),
      }}
    >
      <Column title="Name" dataIndex="name" width={100} />
      <Column title="Key" dataIndex="key" ellipsis={true} width={240} />
      <Column
        title="Action"
        key="x"
        width={60}
        render={(_: unknown, record: { name: string; key: string }) =>
          sshKeys?.length && (
            <>
              <ModalAlert
                headTitle={recordToDelete?.name}
                message="Delete ssh key"
                description="Do you really want to delete this key?"
                type="warning"
                buttons={[
                  <Button
                    key={0}
                    shape="round"
                    className="mr-2 w-24"
                    type="primary"
                    onClick={() => setRecordToDelete(null)} // Close the modal without deleting
                  >
                    Close
                  </Button>,
                  <Button
                    key={1}
                    shape="round"
                    className="ml-2 w-24"
                    color="danger"
                    onClick={() => {
                      if (recordToDelete) {
                        onDeleteKey(recordToDelete) // Safe to call because recordToDelete is not null
                          .then(() => setRecordToDelete(null)) // Close modal on success
                          .catch(_ => null); // Handle error gracefully
                      }
                    }}
                  >
                    Delete
                  </Button>,
                ]}
                show={!!recordToDelete} // Show the modal only when recordToDelete is not null
                setShow={() => setRecordToDelete(null)} // Close the modal on outside click or Close button
              />

              <DeleteOutlined
                onClick={() => setRecordToDelete(record)}
                style={{ color: 'red' }}
              />
            </>
          )
        }
      />
    </Table>
  );
};

export default SSHKeysTable;
