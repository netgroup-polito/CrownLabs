import { FC } from 'react';
import { Typography } from 'antd';

export interface IManagedInstanceHeadingProps {
  tenantId: string;
  tenantDisplayName: string;
  displayName: string;
}

const ManagedInstanceHeading: FC<IManagedInstanceHeadingProps> = ({
  ...props
}) => {
  const { tenantId, tenantDisplayName, displayName } = props;
  const { Text } = Typography;

  return (
    <div className="flex items-center gap-4">
      <Text>{tenantId}</Text>
      <Text>{tenantDisplayName}</Text>
      <Text>{displayName}</Text>
    </div>
  );
};

export default ManagedInstanceHeading;
