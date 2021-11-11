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
  sshKeys?: { name: string; key: string }[];
  onDeleteKey?: (key: {
    name: string;
    key: string;
  }) => boolean | Promise<boolean>;
  onAddKey?: (key: { name: string; key: string }) => boolean | Promise<boolean>;
}

const UserPanel: FC<IUserPanelProps> = props => {
  const { sshKeys, ...otherInfo } = props;
  const [showSSHModal, setShowSSHModal] = useState(false);

  const closeModal = () => setShowSSHModal(false);

  const deleteKey = async (targetKey: { name: string; key: string }) => {
    await props.onDeleteKey?.(targetKey);
  };
  const addKey = async (newKey: { name: string; key: string }) => {
    if (await props.onAddKey?.(newKey)) {
      closeModal();
    }
  };
  const stringHash = (s: string) => {
    return s.split('').reduce((a, b) => {
      a = (a << 5) - a + b.charCodeAt(0);
      return a & a;
    }, 0);
  };

  const generateAvatarUrl = (style: string, seed: string) => {
    return `https://avatars.dicebear.com/api/${style}/${stringHash(seed)}.svg`;
  };

  return (
    <Row className="p-4" align="top">
      <Col xs={24} sm={8} className="text-center">
        <Avatar
          size={100}
          src={generateAvatarUrl('bottts', props.username)}
          icon={<UserOutlined />}
        />
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
            <SSHKeysTable sshKeys={sshKeys} onDeleteKey={deleteKey} />
            <Button className="mt-3" onClick={() => setShowSSHModal(true)}>
              Add SSH key
            </Button>
            <Modal
              title="New SSH key"
              visible={showSSHModal}
              footer={null}
              onCancel={closeModal}
            >
              <SSHKeysForm onSaveKey={addKey} onCancel={closeModal} />
            </Modal>
          </TabPane>
        </Tabs>
      </Col>
    </Row>
  );
};

export default UserPanel;
