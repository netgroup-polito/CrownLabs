import { type FC, useState } from 'react';
import { Row, Col, Avatar, Tabs, Button, Modal } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import UserInfo from '../UserInfo/UserInfo';
import SSHKeysTable from '../SSHKeysTable';
import SSHKeysForm from '../SSHKeysForm';
import { generateAvatarUrl } from '../../../utils';

export interface IUserPanelProps {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  sshKeys?: { name: string; key: string }[];
  onDeleteKey: (key: { name: string; key: string }) => Promise<boolean>;
  onAddKey: (key: { name: string; key: string }) => boolean | Promise<boolean>;
}

const UserPanel: FC<IUserPanelProps> = props => {
  const { sshKeys, onDeleteKey, ...otherInfo } = props;
  const [showSSHModal, setShowSSHModal] = useState(false);

  const closeModal = () => setShowSSHModal(false);

  const addKey = async (newKey: { name: string; key: string }) => {
    if (await props.onAddKey?.(newKey)) {
      closeModal();
    }
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
        <Tabs
          items={[
            {
              key: '1',
              label: 'Info',
              children: <UserInfo {...otherInfo} />,
            },
            {
              key: '2',
              label: 'SSH Keys',
              children: (
                <>
                  <SSHKeysTable sshKeys={sshKeys} onDeleteKey={onDeleteKey} />
                  <Button
                    className="mt-3"
                    onClick={() => setShowSSHModal(true)}
                  >
                    Add SSH key
                  </Button>
                  <Modal
                    title="New SSH key"
                    open={showSSHModal}
                    footer={null}
                    onCancel={closeModal}
                  >
                    <SSHKeysForm onSaveKey={addKey} onCancel={closeModal} />
                  </Modal>
                </>
              ),
            },
          ]}
        ></Tabs>
      </Col>
    </Row>
  );
};

export default UserPanel;
