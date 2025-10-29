import type { FC } from 'react';
import { useQuotaContext } from '../../../contexts/QuotaContext.types';
import QuotaDisplay from '../../workspaces/QuotaDisplay/QuotaDisplay';
import './QuotaStatusBar.less';

const QuotaStatusBar: FC = () => {
  const { consumedQuota, workspaceQuota } = useQuotaContext();

  if (!consumedQuota || !workspaceQuota) return null;

  return (
    <div className="quota-status-bar" style={{ width: '40%' }}>
      <QuotaDisplay
        consumedQuota={consumedQuota}
        workspaceQuota={workspaceQuota}
      />
    </div>
  );
};

export default QuotaStatusBar;
