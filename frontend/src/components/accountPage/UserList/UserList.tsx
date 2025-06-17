import { type FC, useState } from 'react';
import { Table, Modal, Row, Input, Spin, Button } from 'antd';
import {
  UploadOutlined,
  PlusOutlined,
  DeleteOutlined,
  SwapOutlined,
  RedoOutlined,
  CheckOutlined,
} from '@ant-design/icons';
import { Role } from '../../../generated-types';
import { Tooltip } from 'antd';
import UserListFormLogic from '../UserListFormLogic/UserListFormLogic';
import {
  type UserAccountPage,
  type WorkspaceEntry,
  filterUser,
} from '../../../utils';
import UploadProgressModal from '../UploadProgressModal/UploadProgressModal';
import {
  type EnrichedError,
  type SupportedError,
} from '../../../errorHandling/utils';

const { Column } = Table;

export interface IUserListProps {
  onAddUser: (
    users: UserAccountPage[],
    workspaces: WorkspaceEntry[],
  ) => Promise<boolean>;
  onUpdateUser: (user: UserAccountPage, role: Role) => Promise<boolean>;
  setAbortUploading: (value: boolean) => void;
  setUploadingErrors: (errors: EnrichedError[]) => void;
  genericErrorCatcher: (err: SupportedError) => void;
  refreshUserList: () => void;
  abortUploading: boolean;
  users: UserAccountPage[];
  workspaceNamespace: string;
  loading: boolean;
  workspaceName: string;
  uploadedNumber: number;
  uploadedUserNumber: number;
  uploadingErrors: EnrichedError[];
}

const UserList: FC<IUserListProps> = props => {
  const [showUserListModal, setshowUserListModal] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [showUploadModal, setShowUploadModal] = useState(false);
  const closeModal = () => setshowUserListModal(false);

  const handleSearch = (value: string) => {
    setSearchText(value.toLowerCase());
  };

  const handleChangeCurrentRole = (r: UserAccountPage) => {
    props.onUpdateUser(
      r,
      r.currentRole === Role.User ? Role.Manager : Role.User,
    );
  };

  const handleAddUser = async (
    newUser: UserAccountPage,
    workspaces: WorkspaceEntry[],
  ) => await props.onAddUser([newUser], workspaces);

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
                    u1.currentRole?.localeCompare(u2.currentRole!) || 0,
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
            render={(_: unknown, record: UserAccountPage) =>
              props.users.length >= 1 ? (
                <div className="flex justify-center">
                  {record.currentRole === Role.Candidate ? (
                    <Tooltip title="Approve request">
                      <CheckOutlined
                        className="mr-2"
                        onClick={() => handleChangeCurrentRole(record)}
                      />
                    </Tooltip>
                  ) : (
                    <Tooltip title="Swap role">
                      <SwapOutlined
                        className="mr-2"
                        onClick={() => handleChangeCurrentRole(record)}
                      />
                    </Tooltip>
                  )}

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
          destroyOnHidden={true}
          title="Add new User"
          open={showUserListModal}
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
