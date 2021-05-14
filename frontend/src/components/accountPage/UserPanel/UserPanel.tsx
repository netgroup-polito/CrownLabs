import { FC, useState } from 'react';
import { Row, Col, Avatar, Tabs, Button } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import UserInfo from '../UserInfo/UserInfo';
import SSHKeysTable from '../SSHKeysTable';
import Modal from 'antd/lib/modal/Modal';
import SSHKeysForm from '../SSHKeysForm';

const { TabPane } = Tabs;
export interface IUserPanelProps {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  avatar?: string;
  sshKeys?: { name: string; key: string }[];
}

const UserPanel: FC<IUserPanelProps> = props => {
  const { avatar, sshKeys, ...otherInfo } = props;
  const [showSSHModal, setShowSSHModal] = useState(false);

  const closeModal = () => setShowSSHModal(false);

  return (
    <Row className="p-4" align="top">
      <Col xs={24} sm={8} className="text-center">
        <Avatar size="large" icon={avatar ?? <UserOutlined />} />
        <p>
          {otherInfo.firstName} {otherInfo.lastName}
          <br />
          <strong>{otherInfo.username}</strong>
        </p>
      </Col>
      <Col xs={24} sm={16} className="px-4 ">
        <Tabs>
          <TabPane tab="Info" key="1">
            <UserInfo {...otherInfo} />
          </TabPane>
          <TabPane tab="SSH Keys" key="2">
            <SSHKeysTable sshKeys={sshKeys} />
            <Button className="mt-3" onClick={() => setShowSSHModal(true)}>
              Add SSH key
            </Button>
            <Modal
              title="New SSH key"
              visible={showSSHModal}
              footer={null}
              onCancel={closeModal}
            >
              <SSHKeysForm
                onSaveKey={newKey => {
                  closeModal();
                }}
                onCancel={closeModal}
              />
            </Modal>
          </TabPane>
        </Tabs>
      </Col>
    </Row>
  );
};

export default UserPanel;
