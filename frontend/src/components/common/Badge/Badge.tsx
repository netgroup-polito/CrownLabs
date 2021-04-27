import { FC } from 'react';
import { BadgeSize } from '../../../utils';

export interface IBadgeProps {
  value: number;
  size: BadgeSize;
}

const Badge: FC<IBadgeProps> = ({ ...props }) => {
  const { value, size } = props;
  const classPerSize = {
    small: 'h-6 w-6 text-sm ',
    middle: 'h-7 w-7 text-base ',
    large: 'h-8 w-8 text-lg ',
  };
  return (
    <>
      {value ? (
        <span
          className={`
          ${size ? classPerSize[size] : ''}
          mx-2 flex items-center justify-center rounded-lg text-white primary-color-bg`}
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
