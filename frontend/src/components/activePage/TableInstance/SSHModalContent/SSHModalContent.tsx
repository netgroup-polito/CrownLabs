import { FC } from 'react';
import { Table } from 'antd';

export interface ISSHInfo {
  IP: string;
  KEY: string;
}

export interface ISSHModalContentProps {
  sshInfo: ISSHInfo;
}

const SSHModalContent: FC<ISSHModalContentProps> = ({ ...props }) => {
  const { sshInfo } = props;
  const { Column } = Table;
  const data = [];
  for (const [key, val] of Object.entries(sshInfo)) {
    data.push({ heading: key, content: val });
  }

  return (
    <Table dataSource={data} showHeader={false} pagination={false} size="small">
      <Column dataIndex="heading" key="content" />
      <Column dataIndex="content" key="content" />
    </Table>
  );
};

export default SSHModalContent;
