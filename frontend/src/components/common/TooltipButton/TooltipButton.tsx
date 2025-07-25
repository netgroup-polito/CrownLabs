import type { FC } from 'react';
import { Button, Tooltip } from 'antd';
import type { TooltipPlacement } from 'antd/lib/tooltip';
import type { ButtonColorType, ButtonType } from 'antd/lib/button';

export type TooltipButtonData = {
  tooltipTitle?: string;
  tooltipPlacement?: TooltipPlacement;
  icon: React.ReactNode;
  type?: ButtonType;
  color?: ButtonColorType;
};

export interface ITooltipButtonProps {
  TooltipButtonData: TooltipButtonData;
  onClick: () => void;
  className?: string;
}

const TooltipButton: FC<ITooltipButtonProps> = ({ ...props }) => {
  const { onClick, className } = props;

  const { tooltipTitle, tooltipPlacement, icon, type } =
    props.TooltipButtonData;

  return (
    <Tooltip title={tooltipTitle} placement={tooltipPlacement}>
      <Button
        shape="circle"
        type={type}
        size="large"
        icon={icon}
        onClick={onClick}
        className={'flex items-center justify-center ' + className}
      />
    </Tooltip>
  );
};

export default TooltipButton;
