import type { FC } from 'react';
import { Typography, Space, List, Tag, Button } from 'antd';
import { Link } from 'react-router-dom';
import { CodeOutlined } from '@ant-design/icons';
import { Phase2 } from '../../../../generated-types';
import { VITE_APP_CROWNLABS_BASTION_HOST } from '../../../../env';
import type { InstanceEnvironment } from '../../../../utils';

const { Text } = Typography;

export interface ISSHModalContentProps {
  instanceIp: string;
  hasSSHKeys: boolean;
  namespace?: string;
  name?: string;
  prettyName?: string;
  onClose?: () => void;
  environments?: Array<InstanceEnvironment>;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const {
    instanceIp,
    hasSSHKeys,
    environments,
    namespace,
    name /* prettyName, onClose */,
  } = props;

  const getFirstEnvironmentName = () => {
    return environments?.[0]?.name || 'env';
  };

  const buildSSHLink = (envName: string) => {
    if (envName) {
      return `/instance/${namespace}/${name}/${envName}/ssh`;
    }
    return `/instance/${namespace}/${name}/env/ssh`;
  };

  const getEnvironmentStatus = (env: InstanceEnvironment) => {
    const isReady = env.phase === Phase2.Ready;
    return (
      <Tag color={isReady ? 'green' : 'red'}>{env.phase || 'Unknown'}</Tag>
    );
  };

  const getEnvDnsName = (envName?: string) => {
    const eName = envName || getFirstEnvironmentName();
    if (name && namespace) {
      return `${name.replace(/\./g, '-')}-${eName}.${namespace}`;
    }
    return instanceIp;
  };

  const getSshCommand = (envName?: string) => {
    const host = getEnvDnsName(envName);
    return `ssh -J bastion@${VITE_APP_CROWNLABS_BASTION_HOST} crownlabs@${host}`;
  };

  // The browser terminal (WebSSH) uses its own key, injected in every
  // instance, so it is available regardless of the user's SSH keys.
  const renderBrowserButton = (envName: string) => (
    <Link
      to={buildSSHLink(envName)}
      target="_blank"
      rel="noopener noreferrer"
      // onClick={props.onClose}
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
  );

  const isEnvReady = (env: InstanceEnvironment) =>
    !!env.ip && env.phase === Phase2.Ready;

  const hasMultipleEnvs = !!environments && environments.length > 1;

  return (
    <Space
      direction="vertical"
      className="flex justify-center text-center max-w-xl mx-auto"
    >
      <Text className="text-base">
        You can connect to your instance directly from your browser, without any
        additional setup.
      </Text>

      {hasMultipleEnvs ? (
        <List
          dataSource={environments}
          renderItem={env => (
            <List.Item className="flex flex-col">
              <div className="flex justify-between items-center mb-2 gap-2">
                <Text strong>{env.name}</Text>
                {getEnvironmentStatus(env)}
              </div>
              {isEnvReady(env) ? (
                renderBrowserButton(env.name)
              ) : (
                <Text type="secondary" className="text-center">
                  Environment is not ready to connect via SSH
                </Text>
              )}
            </List.Item>
          )}
        />
      ) : (
        renderBrowserButton(getFirstEnvironmentName())
      )}

      <div className="border-t border-gray-400 w-full mt-4" />

      {hasSSHKeys ? (
        <>
          <Text className="mt-5 text-base">
            You have already registered an SSH key. You can also connect from
            your own terminal using:
          </Text>

          {hasMultipleEnvs ? (
            <List
              dataSource={environments?.filter(isEnvReady)}
              locale={{
                emptyText: 'No environment is ready to connect via SSH',
              }}
              renderItem={env => (
                <List.Item className="flex flex-col">
                  <Text strong>{env.name}</Text>
                  <Text type="warning" code copyable className="text-center">
                    {getSshCommand(env.name)}
                  </Text>
                </List.Item>
              )}
            />
          ) : (
            <Text type="warning" code copyable className="flex justify-center">
              {/* FIXME: use netlab username for older VMs, retrieve the correct username
              from the VM's creation timestamp */}
              {getSshCommand()}
            </Text>
          )}
          <Text className="text-sm text-gray-500">
            Want to update your SSH key?
          </Text>
        </>
      ) : (
        <Text className="mt-5 text-base">
          If you would like to access the instance through a standard SSH
          client, you need to upload your public SSH key in your account page.
        </Text>
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
