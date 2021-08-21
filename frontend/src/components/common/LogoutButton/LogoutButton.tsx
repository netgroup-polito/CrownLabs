import { FC } from 'react';
import Button from 'antd-button-color';
import { LogoutOutlined } from '@ant-design/icons';

export interface ILogoutButtonProps {
  logoutHandler: () => void;
  className?: string;
  iconStyle?: React.CSSProperties;
}

const LogoutButton: FC<ILogoutButtonProps> = ({ ...props }) => {
  const { logoutHandler, iconStyle, className } = props;
  return (
    <Button
      onClick={logoutHandler}
      className={
        'm-0 p-0 flex w-auto h-auto items-center bg-transparent border-0 text-red-400 hover:text-red-500 ' +
        className
      }
      size="large"
      type="danger"
      shape={'circle'}
      icon={
        <LogoutOutlined
          className="flex  items-center justify-center "
          style={{ ...iconStyle }}
        />
      }
    />
  );
};

export default LogoutButton;
