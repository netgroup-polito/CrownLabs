import { FC } from 'react';
import { Space, Typography } from 'antd';
import Button from 'antd-button-color';
import { UserSwitchOutlined } from '@ant-design/icons';
import Badge from '../../common/Badge';

const { Text } = Typography;
export interface ITableWorkspaceRowProps {
  text: string;
  nActive: number;
}

const TableWorkspaceRow: FC<ITableWorkspaceRowProps> = ({ ...props }) => {
  const { text, nActive } = props;
  return (
    <div className="w-full flex justify-between">
      <Space size={'middle'}>
        <Badge size="small" value={nActive} className="mx-0" color="green" />
        <Text className="font-bold w-48 sm:w-max" ellipsis>
          {text}
        </Text>
        <Button
          type="ghost"
          shape="circle"
          size="middle"
          className="mr-2"
          icon={<UserSwitchOutlined />}
        />
      </Space>
    </div>
  );
};

export default TableWorkspaceRow;
