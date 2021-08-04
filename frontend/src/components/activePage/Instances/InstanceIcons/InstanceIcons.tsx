import { FC } from 'react';
import { Avatar } from 'antd';
import {
  CodeOutlined,
  DesktopOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  PoweroffOutlined,
} from '@ant-design/icons';

export interface IInstanceIconsProps {
  isGUI: boolean;
  phase: 'ready' | 'creating' | 'failed' | 'stopping' | 'off';
}

const InstanceIcons: FC<IInstanceIconsProps> = ({ ...props }) => {
  const { isGUI, phase } = props;

  const statusIcon = {
    ready: (
      <CheckCircleOutlined className="text-xl text-green-500 hidden lg:inline-block" />
    ),
    failed: (
      <CloseCircleOutlined className="text-xl text-red-500 hidden lg:inline-block" />
    ),
    off: <PoweroffOutlined className="text-xl hidden lg:inline-block" />,
    creating: (
      <LoadingOutlined className="text-xl text-yellow-500 hidden lg:inline-block" />
    ),
    stopping: (
      <LoadingOutlined className="text-xl text-yellow-500 hidden lg:inline-block" />
    ),
  };

  return (
    <div className="flex gap-4 items-center">
      {isGUI ? (
        <DesktopOutlined className="text-xl hidden lg:inline-block" />
      ) : (
        <CodeOutlined className="text-xl hidden lg:inline-block" />
      )}
      {statusIcon[phase]}
      <Avatar shape="square" size={42}>
        VM
      </Avatar>
    </div>
  );
};

export default InstanceIcons;
