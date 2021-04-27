import { FC } from 'react';

export interface IBadgeProps {
  value: number;
}

const Badge: FC<IBadgeProps> = ({ ...props }) => {
  const { value } = props;
  return (
    <>
      {value ? (
        <span
          className="px-2 py-1 text-base rounded-lg text-white bg-blue"
          style={{ backgroundColor: '#1c7afd' }}
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
