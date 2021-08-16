import { FC } from 'react';
import Button, { ButtonType } from 'antd-button-color';
import { Tooltip } from 'antd';
import { TooltipPlacement } from 'antd/lib/tooltip';

export type TooltipButtonData = {
  tooltipTitle?: string;
  tooltipPlacement?: TooltipPlacement;
  icon: React.ReactNode;
  type: ButtonType;
};

export interface ITooltipButtonProps {
  TooltipButtonData: TooltipButtonData;
  onClick: () => void;
  className?: string;
}

const TooltipButton: FC<ITooltipButtonProps> = ({ ...props }) => {
  const { onClick, className } = props;

  const {
    tooltipTitle,
    tooltipPlacement,
    icon,
    type,
  } = props.TooltipButtonData;

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
