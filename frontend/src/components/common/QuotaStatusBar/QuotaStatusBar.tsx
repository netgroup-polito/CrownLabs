import type { FC } from 'react';
import { useQuotaContext } from '../../../contexts/QuotaContext.types';
import QuotaDisplay from '../../workspaces/QuotaDisplay/QuotaDisplay';

const QuotaStatusBar: FC = () => {
  const { consumedQuota, workspaceQuota } = useQuotaContext();

  if (!consumedQuota || !workspaceQuota) return null;

  return (
    <QuotaDisplay
      consumedQuota={consumedQuota}
      workspaceQuota={workspaceQuota}
    />
  );
};

export default QuotaStatusBar;
