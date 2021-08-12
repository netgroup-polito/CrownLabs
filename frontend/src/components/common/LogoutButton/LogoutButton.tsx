import { FC } from 'react';
import Button from 'antd-button-color';
import { LogoutOutlined } from '@ant-design/icons';

export interface ILogoutButtonProps {
  largeMode?: boolean;
  logoutHandler: () => void;
  className?: string;
  iconStyle?: React.CSSProperties;
}

const LogoutButton: FC<ILogoutButtonProps> = ({ ...props }) => {
  const { logoutHandler, largeMode, iconStyle, className } = props;
  return (
    <Button
      onClick={logoutHandler}
      className={
        (largeMode
          ? ''
          : 'm-0 p-0 flex w-auto h-auto items-center bg-transparent justify-center border-0 text-red-400 hover:text-red-500 ') +
        className
      }
      size="large"
      type="danger"
      shape={largeMode ? 'round' : 'circle'}
      icon={
        !largeMode && (
          <LogoutOutlined
            className="flex  items-center justify-center "
            style={{ ...iconStyle }}
          />
        )
      }
    >
      {largeMode && 'Logout'}
    </Button>
  );
};

export default LogoutButton;
