import type { FC } from 'react';
import { Typography, Space } from 'antd';
import { Button } from 'antd';
import { Link } from 'react-router-dom';

const { Text } = Typography;
export interface ISSHModalContentProps {
  instanceIp: string;
  hasSSHKeys: boolean;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { instanceIp, hasSSHKeys } = props;

  return (
    <Space direction="vertical" className="flex justify-center">
      {hasSSHKeys ? (
        <>
          <Text className="flex justify-center uppercase success-color-fg">
            You have registered a SSH key
          </Text>
          <Text className="flex justify-center">
            Connect to your remote instance via the following command:
          </Text>

          <Text type="warning" code copyable className="flex justify-center">
            {/* FIXME: use netlab username for older VMs, retrieve the correct username
            from the VM's creation timestamp */}
            {`ssh -J bastion@ssh.crownlabs.polito.it crownlabs@${instanceIp}`}
          </Text>
        </>
      ) : (
        <>
          <Text className="flex justify-center uppercase danger-color-fg">
            You have not yet registered any SSH key
          </Text>
          <Text className="flex justify-center">
            You need to register a valid SSH KEY before you can use it to
            connect!
          </Text>
          <Text className="flex justify-center">
            Please go to Account page to add a KEY.
          </Text>
          <Text className="flex justify-center">
            <Link to="/account">
              <Button className="mt-3" type="primary" shape="round">
                Go to Account
              </Button>
            </Link>
          </Text>
        </>
      )}
    </Space>
  );
};

export default SSHModalContent;
