import { Button } from 'antd';
import { FC, useState } from 'react';

const ThrowExceptionPage: FC<{}> = ({ ...props }) => {
  const [error, setError] = useState<Error>();
  const onClick = () => {
    setError(new Error('An Uncaught Error'));
  };

  if (error) {
    throw error;
  }

  return (
    <div className="flex w-full h-full justify-center items-center">
      <Button size="large" type="primary" shape="round" onClick={onClick}>
        Trigger Error
      </Button>
    </div>
  );
};

export default ThrowExceptionPage;
