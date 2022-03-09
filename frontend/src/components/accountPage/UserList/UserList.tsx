import { FC, useState } from 'react';
import { Table, Modal, Row, Input, Spin } from 'antd';
import Button from 'antd-button-color';
import {
  UploadOutlined,
  PlusOutlined,
  DeleteOutlined,
  SwapOutlined,
  RedoOutlined,
} from '@ant-design/icons';
import Column from 'antd/lib/table/Column';
import { Role } from '../../../generated-types';
import { Tooltip } from 'antd';
import UserListFormLogic from '../UserListFormLogic/UserListFormLogic';
import { UserAccountPage, filterUser } from '../../../utils';
import UploadProgressModal from '../UploadProgressModal/UploadProgressModal';
import { SupportedError } from '../../../errorHandling/utils';

export interface IUserListProps {
  onAddUser: (users: UserAccountPage[], workspaces: any[]) => Promise<boolean>;
  onUpdateUser: (user: UserAccountPage, role: Role) => Promise<boolean>;
  setAbortUploading: (value: boolean) => void;
  setUploadingErrors: (errors: any[]) => void;
  genericErrorCatcher: (err: SupportedError) => void;
  refreshUserList: () => void;
  abortUploading: boolean;
  users: UserAccountPage[];
  workspaceNamespace: string;
  loading: boolean;
  workspaceName: string;
  uploadedNumber: number;
  uploadedUserNumber: number;
  uploadingErrors: any[];
}

const UserList: FC<IUserListProps> = props => {
  const [showUserListModal, setshowUserListModal] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [showUploadModal, setShowUploadModal] = useState(false);
  const closeModal = () => setshowUserListModal(false);

  const handleSearch = (value: string) => {
    setSearchText(value.toLowerCase());
  };

  const handleChangeCurrentRole = (record: UserAccountPage) => {
    const newRole =
      record.currentRole === Role.Manager ? Role.User : Role.Manager;

    props.onUpdateUser(record, newRole);
  };

  const handleAddUser = async (newUser: UserAccountPage, workspaces: any) =>
    await props.onAddUser([newUser], workspaces);

  return (
    <>
      <Row className="flex justify-center m-4">
        <Input.Search
          placeholder="Search users"
          style={{ width: 300 }}
          onSearch={handleSearch}
          enterButton
          allowClear={true}
        />
      </Row>
      <Spin spinning={props.loading}>
        <Table
          pagination={{ defaultPageSize: 10 }}
          dataSource={
            searchText !== ''
              ? props.users.filter(user => filterUser(user, searchText))
              : props.users.sort(
                  (u1: UserAccountPage, u2: UserAccountPage) =>
                    u1.currentRole?.localeCompare(u2.currentRole!) || 0
                )
          }
          size="small"
        >
          <Column
            title="User ID"
            dataIndex="userid"
            sorter={(a: UserAccountPage, b: UserAccountPage) =>
              a.userid.localeCompare(b.userid)
            }
            key="userid"
            width={170}
          />
          <Column
            responsive={['md', 'lg']}
            title="Name"
            dataIndex="name"
            sorter={(a: UserAccountPage, b: UserAccountPage) =>
              a.name.localeCompare(b.name)
            }
            key="name"
            width={120}
          />
          <Column
            responsive={['md', 'lg']}
            title="Surname"
            dataIndex="surname"
            sorter={(a: UserAccountPage, b: UserAccountPage) =>
              a.surname.localeCompare(b.surname)
            }
            key="surname"
            width={120}
          />
          <Column
            responsive={['sm', 'md', 'lg']}
            title="Email"
            dataIndex="email"
            ellipsis={true}
            key="email"
            width={150}
          />
          <Column
            responsive={['md', 'lg']}
            title={<>Role</>}
            dataIndex="currentRole"
            sorter={(a: UserAccountPage, b: UserAccountPage) =>
              a.currentRole?.localeCompare(b.currentRole!) || 0
            }
            key="currentRole"
            width={80}
          />
          <Column
            title="Action"
            key="x"
            width={60}
            render={(_: any, record: UserAccountPage) =>
              props.users.length >= 1 ? (
                <div className="flex justify-center">
                  <Tooltip title="Swap role">
                    <SwapOutlined
                      className="mr-2"
                      onClick={() => handleChangeCurrentRole(record)}
                    />
                  </Tooltip>

                  <DeleteOutlined className="text-gray-700" />
                </div>
              ) : null
            }
          />
        </Table>
        <Row className="flex justify-between">
          <Row className="flex justify-start mt-4">
            <Button
              type="primary"
              onClick={props.refreshUserList}
              className="m-1"
            >
              {' '}
              Refresh <RedoOutlined />
            </Button>
          </Row>
          <Row className="flex justify-end mt-4">
            <Button
              type="primary"
              onClick={() => setshowUserListModal(true)}
              className="m-1"
            >
              <PlusOutlined />
            </Button>

            <Button
              type="primary"
              className="m-1"
              icon={<UploadOutlined />}
              onClick={() => setShowUploadModal(true)}
            >
              Add from CSV
            </Button>
          </Row>
        </Row>
        <Modal
          destroyOnClose={true}
          title="Add new User"
          visible={showUserListModal}
          footer={null}
          onCancel={closeModal}
        >
          <UserListFormLogic
            onAddUser={handleAddUser}
            onCancel={closeModal}
            workspaceNamespace={props.workspaceNamespace}
            workspaceName={props.workspaceName}
          />
        </Modal>
      </Spin>
      <UploadProgressModal
        onClose={() => setShowUploadModal(false)}
        show={showUploadModal}
        confirmUpload={props.onAddUser}
        workspaceName={props.workspaceName}
        uploadedNumber={props.uploadedNumber}
        setAbortUploading={props.setAbortUploading}
        abortUploading={props.abortUploading}
        uploadingErrors={props.uploadingErrors}
        setUploadingErrors={props.setUploadingErrors}
        uploadedUserNumber={props.uploadedUserNumber}
        genericErrorCatcher={props.genericErrorCatcher}
      />
    </>
  );
};

export default UserList;
