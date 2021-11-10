import { FC } from 'react';
import { Typography, Space } from 'antd';

const { Text } = Typography;
export interface ISSHModalContentProps {
  instanceIp: string;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { instanceIp } = props;

  return (
    <Space direction="vertical">
      <Text>
        You have registered a SSH key, connect to your remote instance via the
        following command:
      </Text>
      <Text code copyable>
        {/* FIXME: use netlab username for older VMs, retrieve the correct username
            from the VM's creation timestamp */}
        {`ssh -J bastion@ssh.crownlabs.polito.it crownlabs@${instanceIp}`}
      </Text>
    </Space>
  );
};

export default SSHModalContent;
