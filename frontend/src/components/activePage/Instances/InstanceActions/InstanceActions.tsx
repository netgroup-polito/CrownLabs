import { FC } from 'react';
import Button from 'antd-button-color';
import {
  DeleteOutlined,
  FolderOpenOutlined,
  MoreOutlined,
} from '@ant-design/icons';

const InstanceActions: FC = () => {
  return (
    <div className="flex justify-end items-center gap-2 pr-2">
      <Button shape="round" className="hidden lg:inline-block">
        SSH
      </Button>
      <Button shape="circle" className="hidden lg:inline-block">
        <FolderOpenOutlined />
      </Button>
      <Button type="primary" shape="round">
        Connect
      </Button>
      <Button type="danger" shape="circle">
        <DeleteOutlined />
      </Button>
      <MoreOutlined className="lg:hidden" />
    </div>
  );
};

export default InstanceActions;
