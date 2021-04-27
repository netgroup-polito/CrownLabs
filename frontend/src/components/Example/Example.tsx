import { FC } from 'react';
import { Button } from 'antd';
import { SizeType } from 'antd/lib/config-provider/SizeContext';
import './Example.css';

export interface IExampleProps {
  text: string;
  disabled: boolean;
  onClick: () => void;
  specialCSS: boolean;
  size: SizeType;
}

const Example: FC<IExampleProps> = ({ ...props }) => {
  const { text, disabled, onClick, size, specialCSS } = props;
  return (
    <Button
      disabled={disabled}
      size={size}
      onClick={onClick}
      type="primary"
      className="m-10"
    >
      <h5 className={specialCSS ? 'rainbow-text' : ''}>{text}</h5>
    </Button>
  );
};

export default Example;
