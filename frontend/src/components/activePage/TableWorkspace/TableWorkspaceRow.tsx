import { FC } from 'react';
import { Space, Typography } from 'antd';
import Button from 'antd-button-color';
import { UserSwitchOutlined } from '@ant-design/icons';
import Badge from '../../common/Badge';

const { Text } = Typography;
export interface ITableWorkspaceRowProps {
  title: string;
  nActive: number;
  id: string;
  expandRow: (rowId: string) => void;
}

const TableWorkspaceRow: FC<ITableWorkspaceRowProps> = ({ ...props }) => {
  const { title, nActive, id, expandRow } = props;
  return (
    <div
      className="w-full flex justify-between cursor-pointer"
      onClick={() => expandRow(id)}
    >
      <Space size="middle">
        <Badge size="small" value={nActive} className="mx-0" color="green" />
        <Text className="font-bold w-48 xs:w-56 sm:w-max" ellipsis>
          {title}
        </Text>
        <Button
          disabled={true}
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
