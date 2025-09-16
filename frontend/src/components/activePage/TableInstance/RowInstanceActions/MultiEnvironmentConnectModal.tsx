import { type FC } from 'react';
import { Modal, List, Button, Typography, Tag } from 'antd';
import { Phase } from '../../../../generated-types';
import type { InstanceEnvironment } from '../../../../utils';

const { Text } = Typography;

interface IMultiEnvironmentConnectModalProps {
  open: boolean;
  onCancel: () => void;
  environments: InstanceEnvironment[];
  instanceUrl: string | null;
  gui: boolean;
  setSshModal: (show: boolean) => void;
}

const MultiEnvironmentConnectModal: FC<IMultiEnvironmentConnectModalProps> = ({
  open,
  onCancel,
  environments,
  instanceUrl,
  gui,
  setSshModal,
}) => {
  const isEnvironmentReady = (env: InstanceEnvironment) => env.phase === Phase.Ready;

  const handleEnvironmentConnect = (env: InstanceEnvironment) => {
    if (gui) {
      if (!instanceUrl) {
        onCancel();
        return;
      }
      const envUrl = `${instanceUrl}${env.name}`;
      window.open(envUrl, '_blank');
    } else {
      setSshModal(true);
    }
    onCancel();
  };

  const getEnvironmentStatus = (env: InstanceEnvironment) => {
    const isReady = isEnvironmentReady(env);
    return (
      <Tag color={isReady ? 'green' : 'red'}>
        {env.phase || 'Unknown'}
      </Tag>
    );
  };

  return (
    <Modal
      title="Select Environment to Connect"
      open={open}
      onCancel={onCancel}
      footer={null}
    >
      <List
        dataSource={environments}
        renderItem={(env) => (
          <List.Item
            actions={[
              <Button
                key="connect"
                type="primary"
                shape="round"
                disabled={!isEnvironmentReady(env)}
                onClick={() => handleEnvironmentConnect(env)}
              >
                {gui ? 'Connect' : 'SSH'}
              </Button>
            ]}
          >
            <List.Item.Meta
              title={
                <div className="flex items-center gap-2">
                  <Text strong>{env.name}</Text>
                  {getEnvironmentStatus(env)}
                </div>
              }
              description={
                <div>
                  <Text type="secondary">
                    {env.ip ? `IP: ${env.ip}` : 'IP not assigned'}
                  </Text>
                  {!isEnvironmentReady(env) && (
                    <div>
                      <Text type="warning">
                        Environment is not ready to connect
                      </Text>
                    </div>
                  )}
                </div>
              }
            />
          </List.Item>
        )}
      />
    </Modal>
  );
};

export default MultiEnvironmentConnectModal;