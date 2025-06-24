import type { FC } from 'react';
import { Typography, Space, Button } from 'antd';
import { Link } from 'react-router-dom';
import { CodeOutlined } from '@ant-design/icons';
const { Text } = Typography;

export interface ISSHModalContentProps {
  instanceIp: string;
  hasSSHKeys: boolean;
  namespace?: string;
  name?: string;
  prettyName?: string;
  onClose?: () => void;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { instanceIp, hasSSHKeys } = props;

  const ENV_PLACEHOLDER = 'env';

  return (
    <Space
      direction="vertical"
      className="flex justify-center text-center max-w-xl mx-auto"
    >
      <Text className="text-base">
        You can open the terminal in your browser, or set up a personal SSH key
        to use your own terminal.
      </Text>

      <Link
        to={`/instance/${props.namespace}/${props.name}/${ENV_PLACEHOLDER}/ssh`}
        target="_blank"
        rel="noopener noreferrer"
        onClick={props.onClose}
      >
        <Button
          className="mt-4 bg-green-600 hover:bg-green-700"
          type="primary"
          shape="round"
        >
          <CodeOutlined></CodeOutlined>
          Connect via browser
        </Button>
      </Link>

      <div className="border-t border-gray-400 w-full mt-4" />

      {hasSSHKeys ? (
        <>
          <Text className="mt-5 text-base">
            You have already registered an SSH key. You can connect via terminal
            using:
          </Text>
          <Text type="warning" code copyable>
            ssh -J bastion@ssh.crownlabs.polito.it crownlabs@{instanceIp}
          </Text>
          <Text className="text-sm text-gray-500">
            Want to update your SSH key?
          </Text>
        </>
      ) : (
        <>
          <Text className="mt-5  danger-color-fg font-semibold">
            To connect via terminal, you need to register a personal SSH key.
          </Text>
        </>
      )}

      <Link to="/account">
        <Button className="mt-2" type="default" shape="round">
          Go to Account
        </Button>
      </Link>
    </Space>
  );
};

export default SSHModalContent;
