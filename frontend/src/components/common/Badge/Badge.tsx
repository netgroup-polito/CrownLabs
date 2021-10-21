import { FC } from 'react';
import { BadgeSize } from '../../../utils';

export interface IBadgeProps {
  value: number;
  size: BadgeSize;
  className?: string;
  color?: 'yellow' | 'blue' | 'green';
}

const Badge: FC<IBadgeProps> = ({ ...props }) => {
  const { value, size, className, color } = props;
  const classPerSize = {
    small: 'h-6 px-1 text-sm ',
    middle: 'h-7 w-7 text-base ',
    large: 'h-8 w-8 text-lg ',
  };

  const colorByProps = {
    green: 'success-color-bg',
    yellow: 'warning-color-bg ',
    blue: 'primary-color-bg ',
  };
  return (
    <>
      {value ? (
        <span
          style={{ minWidth: '1.5rem' }}
          className={`
          ${size ? classPerSize[size] : ''}
          ${className} flex items-center justify-center rounded-lg text-white ${
            color ? colorByProps[color] : 'primary-color-bg'
          }`}
        >
          {value}
        </span>
      ) : (
        ''
      )}
    </>
  );
};

export default Badge;
