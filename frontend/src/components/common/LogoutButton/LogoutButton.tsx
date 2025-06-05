import type { FC } from 'react';
import { Button } from 'antd';
import { LogoutOutlined } from '@ant-design/icons';
import { Tooltip } from 'antd';

export interface ILogoutButtonProps {
  logoutHandler: () => void;
  className?: string;
  iconStyle?: React.CSSProperties;
  buttonStyle?: React.CSSProperties;
}

const LogoutButton: FC<ILogoutButtonProps> = ({ ...props }) => {
  const { logoutHandler, iconStyle, className, buttonStyle } = props;
  return (
    <Button
      onClick={logoutHandler}
      style={{ ...buttonStyle }}
      className={
        'm-0 p-0 flex h-auto items-center bg-transparent border-0 text-red-400 hover:text-red-500 ' +
        className
      }
      size="large"
      color="red"
      shape={'circle'}
      icon={
        <Tooltip trigger="hover" placement="bottom" title="Logout">
          <LogoutOutlined
            className="flex  items-center justify-center "
            style={{ ...iconStyle }}
          />
        </Tooltip>
      }
    />
  );
};

export default LogoutButton;
