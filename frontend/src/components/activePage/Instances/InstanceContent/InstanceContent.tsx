import { FC } from 'react';
import { Button } from 'antd';
import { InfoOutlined } from '@ant-design/icons';
import VmInstanceActions from '../InstanceActions/InstanceActions';
import ManagedInstanceHeading from './ManagedInstanceHeading';

export interface IInstanceContentProps {
  isManaged: boolean;
  displayName: string;
  tenantId: string;
  tenantDisplayName: string;
}

const InstanceContent: FC<IInstanceContentProps> = ({ ...props }) => {
  const { displayName, tenantId, tenantDisplayName, isManaged } = props;
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
      <Button shape="circle">
        <InfoOutlined />
      </Button>
      <VmInstanceActions />
    </div>
  );
};

export default InstanceContent;
