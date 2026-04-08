import { Empty, Spin } from 'antd';
import { type FC, useContext, useMemo, useState } from 'react';
import { TenantContext } from '../../../contexts/TenantContext';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import type { WorkspaceRole } from '../../../utils';
import { type Instance, type User } from '../../../utils';
import { sorter } from '../../../utilsLogic';
import TableInstance from './TableInstance';
import './TableInstance.less';
export interface ITableInstanceLogicProps {
  viewMode: WorkspaceRole;
  showGuiIcon: boolean;
  extended: boolean;
  user: User;
}

const TableInstanceLogic: FC<ITableInstanceLogicProps> = ({ ...props }) => {
  const { viewMode, extended, showGuiIcon } = props;
  const { hasSSHKeys } = useContext(TenantContext);
  const {
    instances: allInstances,
    loading,
    error,
  } = useContext(OwnedInstancesContext);

  const [sortingData, setSortingData] = useState<{
    sortingType: string;
    sorting: number;
  }>({ sortingType: '', sorting: 0 });

  const handleSorting = (sortingType: string, sorting: number) => {
    setSortingData({ sortingType, sorting });
  };

  // Sort instances based on current sorting settings
  const instances = useMemo(() => {
    return [...allInstances].sort((a, b) =>
      sorter(
        a,
        b,
        sortingData.sortingType as keyof Instance,
        sortingData.sorting,
      ),
    );
  }, [allInstances, sortingData]);

  return (
    <>
      {!loading && !error ? (
        instances.length ? (
          <TableInstance
            showGuiIcon={showGuiIcon}
            viewMode={viewMode}
            hasSSHKeys={hasSSHKeys}
            instances={instances}
            extended={extended}
            handleSorting={handleSorting}
            showAdvanced={true}
          />
        ) : (
          <div className="w-full h-full flex-grow flex flex-wrap content-center justify-center py-5 ">
            <div className="w-full pb-10 flex justify-center">
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={false} />
            </div>
            <p className="text-xl xs:text-3xl text-center px-5 xs:px-24">
              No running instances
            </p>
          </div>
        )
      ) : (
        <div className="flex justify-center h-full items-center">
          {loading ? (
            <Spin size="large" spinning={loading} />
          ) : (
            <>{error && <p>{error.message}</p>}</>
          )}
        </div>
      )}
    </>
  );
};

export default TableInstanceLogic;
