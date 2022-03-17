import { DeleteOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import Button from 'antd-button-color';
import Column from 'antd/lib/table/Column';
import { FC, useState } from 'react';
import { ModalAlert } from '../../common/ModalAlert';

export interface ISSHKeysTableProps {
  sshKeys?: { name: string; key: string }[];
  onDeleteKey: (key: { name: string; key: string }) => Promise<boolean>;
}

const SSHKeysTable: FC<ISSHKeysTableProps> = props => {
  const { sshKeys, onDeleteKey } = props;
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);
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
        render={(_: any, record: { name: string; key: string }) =>
          sshKeys?.length && (
            <>
              <ModalAlert
                headTitle={record.name}
                message="Delete ssh key"
                description="Do you really want to delete this key?"
                type="warning"
                buttons={[
                  <Button
                    key={0}
                    shape="round"
                    className="mr-2 w-24"
                    type="primary"
                    onClick={() => setShowDeleteModalConfirm(false)}
                  >
                    Close
                  </Button>,
                  <Button
                    key={1}
                    shape="round"
                    className="ml-2 w-24"
                    type="danger"
                    onClick={() =>
                      onDeleteKey(record)
                        .then(() => setShowDeleteModalConfirm(false))
                        .catch(err => null)
                    }
                  >
                    Delete
                  </Button>,
                ]}
                show={showDeleteModalConfirm}
                setShow={setShowDeleteModalConfirm}
              />

              <DeleteOutlined
                onClick={() => setShowDeleteModalConfirm(true)}
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
