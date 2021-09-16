import { FC, useState } from 'react';
import { Table, Button, Upload, Modal, Row, Typography, Input } from 'antd';
import {
  UploadOutlined,
  PlusOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import Column from 'antd/lib/table/Column';
import Papa from 'papaparse';
import UserListForm from '../UserListForm';
import ResetButton from './ResetButton';

const Text = Typography.Text;

export type User = {
  key: Number;
  userid: string;
  name: string;
  surname: string;
  email: string;
};

export interface IUserListProps {
  onAddUser: (users: User[]) => void;
}

const UserList: FC<IUserListProps> = props => {
  const [showUserListModal, setshowUserListModal] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [fileError, setFileError] = useState<string>('');

  const closeModal = () => setshowUserListModal(false);

  const onCsvUploaded = (fileInfo: any) => {
    if (fileInfo.file.status === 'removed') {
      setUsers([]);
      return;
    }

    Papa.parse<any>(fileInfo.file, {
      header: true,
      skipEmptyLines: true,
      complete: (result, _) => {
        for (const line of result.data) {
          if (
            !line['NOME'] ||
            !(line['COGNOME - (*) Inserito dal docente'] || line['COGNOME']) ||
            !line['MATRICOLA'] ||
            !line['EMAIL']
          ) {
            setFileError(
              'Invalid file format, must contain <MATRICOLA, NOME, COGNOME (o COGNOME - (*) Inserito dal docente), EMAIL>'
            );
            return;
          }
        }
        setUsers(() => {
          const users = result.data.map((user: any, index: Number) => {
            return {
              name: user['NOME'] ?? '',
              surname:
                user['COGNOME - (*) Inserito dal docente'] ??
                user['COGNOME'] ??
                '',
              userid: user['MATRICOLA'] ?? '',
              email: user['EMAIL'] ?? '',
              key: index,
            };
          });
          props.onAddUser(users);
          return users;
        });

        setFileError('');
      },
    });
  };

  function makeFilter(field: string) {
    return (value: any, record: any) =>
      record[field]
        ? record[field].toString().toLowerCase().includes(value.toLowerCase())
        : '';
  }

  return (
    <>
      <Table dataSource={users}>
        <Column
          title="User ID"
          dataIndex="userid"
          sorter={(a: User, b: User) => a.userid.localeCompare(b.userid)}
          key="userid"
          filterIcon={<SearchOutlined />}
          onFilter={makeFilter('userid')}
          filterDropdown={({
            setSelectedKeys,
            selectedKeys,
            confirm,
            clearFilters,
          }) => (
            <div style={{ padding: 8 }}>
              <Input
                placeholder="Search by Userid"
                onChange={e =>
                  setSelectedKeys(e.target.value ? [e.target.value] : [])
                }
                value={selectedKeys[0]}
                onPressEnter={() => confirm()}
              />
              <ResetButton onClick={clearFilters} />
            </div>
          )}
          width={120}
        />
        <Column
          title="Name"
          dataIndex="name"
          sorter={(a: User, b: User) => a.name.localeCompare(b.name)}
          key="name"
          filterIcon={<SearchOutlined />}
          onFilter={makeFilter('name')}
          filterDropdown={({
            setSelectedKeys,
            selectedKeys,
            confirm,
            clearFilters,
          }) => (
            <div style={{ padding: 8 }}>
              <Input
                placeholder="Search by Name"
                onChange={e =>
                  setSelectedKeys(e.target.value ? [e.target.value] : [])
                }
                value={selectedKeys[0]}
                onPressEnter={() => confirm()}
              />
              <ResetButton onClick={clearFilters} />
            </div>
          )}
          width={120}
        />
        <Column
          title="Surname"
          dataIndex="surname"
          sorter={(a: User, b: User) => a.surname.localeCompare(b.surname)}
          key="surname"
          filterIcon={<SearchOutlined />}
          onFilter={makeFilter('surname')}
          filterDropdown={({
            setSelectedKeys,
            selectedKeys,
            confirm,
            clearFilters,
          }) => (
            <div style={{ padding: 8 }}>
              <Input
                placeholder="Search by Surname"
                onChange={e =>
                  setSelectedKeys(e.target.value ? [e.target.value] : [])
                }
                value={selectedKeys[0]}
                onPressEnter={() => confirm()}
              />
              <ResetButton onClick={clearFilters} />
            </div>
          )}
          width={120}
        />
        <Column
          title="Email"
          dataIndex="email"
          filterIcon={<SearchOutlined />}
          onFilter={makeFilter('email')}
          filterDropdown={({
            setSelectedKeys,
            selectedKeys,
            confirm,
            clearFilters,
          }) => (
            <div style={{ padding: 8 }}>
              <Input
                placeholder="Search by Email"
                onChange={e =>
                  setSelectedKeys(e.target.value ? [e.target.value] : [])
                }
                value={selectedKeys[0]}
                onPressEnter={() => confirm()}
              />
              <ResetButton onClick={clearFilters} />
            </div>
          )}
          key="email"
          width={120}
        />
      </Table>
      <Row className="flex justify-end mt-4">
        <Button
          type="primary"
          onClick={() => setshowUserListModal(true)}
          className="m-1"
        >
          <PlusOutlined />
        </Button>
        <Upload
          name="file"
          beforeUpload={() => false}
          accept=".csv"
          onChange={onCsvUploaded}
          fileList={[]}
          maxCount={1}
        >
          <Button type="primary" className="m-1" icon={<UploadOutlined />}>
            Add many from CSV
          </Button>
        </Upload>
        {fileError && (
          <Text className="m-2" type="danger">
            {fileError}
          </Text>
        )}
      </Row>
      <Modal
        title="Add new User"
        visible={showUserListModal}
        footer={null}
        onCancel={closeModal}
      >
        <UserListForm
          onAddUser={newUser => {
            setUsers(users => [...users, newUser]);
            props.onAddUser([newUser]);
            closeModal();
          }}
          onCancel={closeModal}
        />
      </Modal>
    </>
  );
};

export default UserList;
