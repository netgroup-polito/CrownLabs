import { FC } from 'react';
import { Button } from 'antd';
import { SizeType } from 'antd/lib/config-provider/SizeContext';
import './ExampleButton.css';

export interface IExampleButtonProps {
  text: string;
  disabled: boolean;
  onClick: () => void;
  specialCSS: boolean;
  size: SizeType;
}

const Example: FC<IExampleButtonProps> = ({ ...props }) => {
  const { text, disabled, onClick, size, specialCSS } = props;
  return (
    <div className="p-10">
      <Button
        disabled={disabled}
        size={size}
        onClick={onClick}
        type="primary"
        className="m-10 rounded-xl"
      >
        <h5 className={specialCSS ? 'rainbow-text' : ''}>{text}</h5>
      </Button>
    </div>
  );
};

export default Example;
