import type { FC } from 'react';
import { Typography, Space, Button } from 'antd';
import { Link } from 'react-router-dom';

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

  return (
    <Space direction="vertical" className="flex justify-center text-center max-w-xl mx-auto">
      <Text className="text-base">
        Use your browser to open the terminal now, or configure a personal SSH key to access it from your terminal.
      </Text>

      <Button
        className="mt-4"
        type="primary"
        shape="round"
        onClick={() => {
          window.open(
            `/ssh/${props.namespace}/${props.name}?prettyName=${encodeURIComponent(
              props.prettyName ?? ''
            )}`,
            '_blank',
            'noopener,noreferrer'
          );
          props.onClose?.();
        }}
      >
        Connect via browser
      </Button>

      {hasSSHKeys ? (
        <>
          <Text className="mt-5 text-base">
            You have already registered an SSH key. You can connect via terminal using:
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
