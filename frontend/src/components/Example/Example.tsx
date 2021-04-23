import { FC } from 'react';

export interface IExampleProps {
  text: string;
  disabled: boolean;
  onClick: () => void;
}

const Example: FC<IExampleProps> = ({ ...props }) => {
  const { text, disabled, onClick } = props;
  return (
    <button
      style={{ color: 'blue', fontSize: '3rem', opacity: disabled ? 0.5 : 1 }}
      onClick={onClick}
    >
      {text}
    </button>
  );
};

export default Example;
