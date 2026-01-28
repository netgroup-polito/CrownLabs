import { type FC } from 'react';
import { Button } from 'antd';

export interface IDriveButtonProps {
  onClick: () => void;
  className?: string;
  size?: 'small' | 'middle' | 'large';
  type?: 'default' | 'primary' | 'text' | 'link';
  shape?: 'default' | 'circle' | 'round';
  ghost?: boolean;
}

const DriveButton: FC<IDriveButtonProps> = ({
  onClick,
  className = '',
  size = 'large',
  type = 'default',
  shape = 'round',
  ghost = false,
}) => {
  return (
    <Button
      onClick={onClick}
      className={className}
      size={size}
      type={type}
      shape={shape}
      ghost={ghost}
    >
      Drive
    </Button>
  );
};

export default DriveButton;
