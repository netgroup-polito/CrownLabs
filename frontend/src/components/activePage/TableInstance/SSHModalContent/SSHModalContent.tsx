import type { FC } from 'react';
import { Typography, Space, List, Tag, Button } from 'antd';
import { Link } from 'react-router-dom';
import { Phase } from '../../../../generated-types';
import type { InstanceEnvironment } from '../../../../utils';

const { Text } = Typography;
export interface ISSHModalContentProps {
  instanceIp: string;
  hasSSHKeys: boolean;
  environments?: Array<{
    name: string;
    ip?: string;
    phase?: Phase;
    guiEnabled?: boolean;
  }>;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { instanceIp, hasSSHKeys, environments} = props;

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
    <Space direction="vertical" className="flex justify-center">
      {hasSSHKeys ? (
        <>
          <Text className="flex justify-center uppercase success-color-fg">
            You have registered a SSH key
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
