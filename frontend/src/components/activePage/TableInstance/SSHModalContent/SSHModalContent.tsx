import type { FC } from 'react';
import { Typography, Space, List, Tag, Button } from 'antd';
import { Link } from 'react-router-dom';
import { CodeOutlined } from '@ant-design/icons';
import { Phase } from '../../../../generated-types';
import type { InstanceEnvironment } from '../../../../utils';

const { Text } = Typography;

export interface ISSHModalContentProps {
  instanceIp: string;
  hasSSHKeys: boolean;
  namespace?: string;
  name?: string;
  prettyName?: string;
  onClose?: () => void;
  environments?: Array<{
    name: string;
    ip?: string;
    phase?: Phase;
    guiEnabled?: boolean;
  }>;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { instanceIp, hasSSHKeys, environments, namespace, name, prettyName, onClose } = props;

  const getEnvironmentStatus = (env: InstanceEnvironment) => {
    const isReady = env.phase === Phase.Ready;
    return (
      <Tag color={isReady ? 'green' : 'red'}>
        {env.phase || 'Unknown'}
      </Tag>
    );
  };

  const getSshCommand = (envIP: string) => {
    return `ssh -J bastion@ssh.crownlabs.polito.it crownlabs@${envIP}`;
  };

  return (
    <Space
      direction="vertical"
      className="flex justify-center text-center max-w-xl mx-auto"
    >
      <Text className="text-base">
        You can open the terminal in your browser, or set up a personal SSH key
        to use your own terminal.
      </Text>

      {namespace && name && (
        <Link
          to={`/instance/${namespace}/${name}/${environments?.[0]?.name || 'env'}/ssh`}
          target="_blank"
          rel="noopener noreferrer"
          onClick={onClose}
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
      )}

      <div className="border-t border-gray-400 w-full mt-4" />

      {hasSSHKeys ? (
        <>
          <Text className="mt-5 text-base">
            You have already registered an SSH key. You can connect via terminal
            using:
          </Text>
          
          {environments && environments.length > 1 ? (
            <>
              <Text className="flex justify-center mb-2">
                Connect to your remote environments via the following commands:
              </Text>
              
              <List
                dataSource={environments}
                renderItem={(env) => (
                  <List.Item className="flex flex-col">
                    <div className="flex justify-between items-center mb-2 gap-2">
                      <Text strong>{env.name}</Text>
                      {getEnvironmentStatus(env)}
                    </div>
                    {env.ip && env.phase === Phase.Ready ? (
                      <Text type="warning" code copyable className="text-center">
                        {getSshCommand(env.ip)}
                      </Text>
                    ) : (
                      <Text type="secondary" className="text-center">
                        Environment is not ready to connect via SSH
                      </Text>
                    )}
                  </List.Item>
                )}
              />
            </>
          ) : (
            <>
              <Text className="flex justify-center">
                Connect to your remote instance via the following command:
              </Text>

              <Text type="warning" code copyable className="flex justify-center">
                {/* FIXME: use netlab username for older VMs, retrieve the correct username
                from the VM's creation timestamp */}
                {getSshCommand(instanceIp)}
              </Text>
            </>
          )}
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
