import { Button } from 'antd';
import type { FC } from 'react';
import { useContext } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';

const ThrowExceptionPage: FC = () => {
  const { makeErrorCatcher } = useContext(ErrorContext);
  const renderErrorCatcher = makeErrorCatcher(ErrorTypes.RenderError);
  const triggerError = (n: number) => {
    const errQueue = [];
    for (let index = 0; index < n; index++) {
      const err = new Error('DEBUG: error test ' + index);
      errQueue.push(err);
      renderErrorCatcher(err);
    }
  };
  const triggerSameError = (n: number) => {
    const errQueue = [];
    for (let index = 0; index < n; index++) {
      const err = new Error('DEBUG: error test');
      errQueue.push(err);
      renderErrorCatcher(err);
    }
  };

  return (
    <div className="flex w-full h-full flex-wrap items-center pb-12">
      {[1, 2, 3].map(c => (
        <div key={c} className="flex w-full justify-center">
          <Button
            size="large"
            type="primary"
            shape="round"
            onClick={() => triggerError(c)}
          >
            Trigger {c} Errors
          </Button>
        </div>
      ))}
      <div key={5} className="flex w-full justify-center">
        <Button
          size="large"
          type="primary"
          shape="round"
          onClick={() => triggerSameError(5)}
        >
          Trigger 5 identical Errors
        </Button>
      </div>
    </div>
  );
};

export default ThrowExceptionPage;
