import { FC } from 'react';
import { Popover, Tooltip } from 'antd';
import Button from 'antd-button-color';
import {
  ExclamationCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { Instance } from '../../../../utils';

export interface IRowInstanceActionsPersistentProps {
  extended: boolean;
  instance: Instance;
  startInstance: (idInstance: number, idTemplate: string) => void;
  stopInstance: (idInstance: number, idTemplate: string) => void;
}

const RowInstanceActionsPersistent: FC<IRowInstanceActionsPersistentProps> = ({
  ...props
}) => {
  const { extended, instance, startInstance, stopInstance } = props;

  const { status, id, idTemplate } = instance;

  const font22px = { fontSize: '22px' };

  return status === 'VmiReady' ? (
    <Tooltip placement="top" title={'Pause'}>
      <Button
        className={`hidden ${
          extended ? 'sm:block' : 'xs:block'
        } flex items-center`}
        type="warning"
        with="link"
        shape="circle"
        size="middle"
        icon={
          <PauseCircleOutlined
            className={'flex justify-center items-center'}
            style={font22px}
          />
        }
        onClick={() => startInstance(id, idTemplate!)}
      />
    </Tooltip>
  ) : status === 'VmiOff' ? (
    <Tooltip placement="top" title={'Start'}>
      <Button
        className={`hidden ${extended ? 'sm:block' : 'xs:block'} py-0`}
        type="success"
        with="link"
        shape="circle"
        size="middle"
        icon={
          <PlayCircleOutlined
            className={'flex justify-center items-center'}
            style={font22px}
          />
        }
        onClick={() => stopInstance(id, idTemplate!)}
      />
    </Tooltip>
  ) : (
    <Popover
      placement="top"
      title={'No Actions Available'}
      content={'Current instance Status: ' + status}
    >
      <div className="cursor-not-allowed">
        <Button
          className={`hidden pointer-events-none ${
            extended ? 'sm:block' : 'xs:block'
          } py-0`}
          type="primary"
          with="link"
          shape="circle"
          size="middle"
          disabled={true}
          icon={
            <ExclamationCircleOutlined
              className={'flex justify-center items-center'}
              style={font22px}
            />
          }
          onClick={() => stopInstance(id, idTemplate!)}
        />
      </div>
    </Popover>
  );
};

export default RowInstanceActionsPersistent;
