import { FC, Dispatch, SetStateAction } from 'react';
import { Button, Popover, List } from 'antd';
import { InfoOutlined } from '@ant-design/icons';
import InstanceActions from '../InstanceActions/InstanceActions';
import { ISSHInfo } from '../InstancesTable/InstancesTable';
import ManagedInstanceHeading from './ManagedInstanceHeading';

export interface IInstanceContentProps {
  isManaged: boolean;
  displayName: string;
  tenantId: string;
  tenantDisplayName: string;
  ip: string;
  phase: 'ready' | 'creating' | 'failed' | 'stopping' | 'off';
  toggleModal: Dispatch<SetStateAction<boolean>>;
  setSshInfo: Dispatch<SetStateAction<ISSHInfo>>;
}

const InstanceContent: FC<IInstanceContentProps> = ({ ...props }) => {
  const {
    displayName,
    tenantId,
    tenantDisplayName,
    isManaged,
    ip,
    phase,
    toggleModal,
    setSshInfo,
  } = props;
  return (
    <div className="flex justify-between items-center">
      {isManaged ? (
        <ManagedInstanceHeading
          displayName={displayName}
          tenantId={tenantId}
          tenantDisplayName={tenantDisplayName}
        />
      ) : (
        displayName
      )}
      <Popover
        placement="bottom"
        trigger={['click']}
        content={
          <List
            size="small"
            className="p-0"
            dataSource={['IP: 192.168.1.1', 'Text 2', 'Text 3']}
            renderItem={item => <List.Item className="px-0">{item}</List.Item>}
          />
        }
      >
        <Button shape="circle" disabled={phase !== 'ready'}>
          <InfoOutlined />
        </Button>
      </Popover>
      <InstanceActions
        setSshInfo={setSshInfo}
        phase={phase}
        toggleModal={toggleModal}
        ip={ip}
      />
    </div>
  );
};

export default InstanceContent;
