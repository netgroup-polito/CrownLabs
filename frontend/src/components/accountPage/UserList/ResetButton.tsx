import type { FC, MouseEventHandler } from 'react';
import { Button, Space } from 'antd';
export interface ResetButtonProps {
  onClick: MouseEventHandler<Element> | undefined;
}

const ResetButton: FC<ResetButtonProps> = ({ ...props }) => {
  const { onClick } = props;
  return (
    <Space>
      <Button onClick={onClick} size="small" style={{ width: 90 }}>
        Reset
      </Button>
    </Space>
  );
};

export default ResetButton;
