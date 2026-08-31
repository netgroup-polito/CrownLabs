import type { Dispatch, FC, SetStateAction } from 'react';
import { useContext, useEffect, useMemo } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { WorkspaceRole, type Instance, type Workspace } from '../../../utils';
import './TableWorkspace.less';
import TableInstance from '../TableInstance/TableInstance';
import { TenantContext } from '../../../contexts/TenantContext';

export interface ITableWorkspaceProps {
  instances: Array<Instance>;
  workspaces: Array<Workspace>;
  showAdvanced: boolean;
  showCheckbox: boolean;
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string,
  ) => void;
  destroySelectedTrigger: boolean;
  setDestroySelectedTrigger: Dispatch<SetStateAction<boolean>>;
  setSelectedPersistent: Dispatch<SetStateAction<boolean>>;
  selectiveDestroy: string[];
  selectToDestroy: (instanceId: string) => void;
}

const TableWorkspace: FC<ITableWorkspaceProps> = ({ ...props }) => {
  const {
    instances,
    workspaces,
    showAdvanced,
    showCheckbox,
    handleManagerSorting,
    destroySelectedTrigger,
    setDestroySelectedTrigger,
    selectiveDestroy,
    selectToDestroy,
    setSelectedPersistent,
  } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const { hasSSHKeys } = useContext(TenantContext);

  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const destroySelected = async () => {
    const selection = instances.filter(i => selectiveDestroy.includes(i.id));
    for (const { tenantNamespace, name: instanceId, id } of selection) {
      await deleteInstanceMutation({
        variables: { tenantNamespace, instanceId },
      });
      // Removing from selection list after deletion
      selectToDestroy(id);
    }
  };

  useEffect(() => {
    const persistent =
      (instances &&
        instances.filter(
          i => selectiveDestroy.includes(i.id) && i.persistent,
        )) ||
      [];
    setSelectedPersistent(persistent.length > 0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectiveDestroy]);

  useEffect(() => {
    if (destroySelectedTrigger) {
      setDestroySelectedTrigger(false);
      destroySelected();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [destroySelectedTrigger]);

  const workspacePrettyName = useMemo(
    () => Object.fromEntries(workspaces.map(ws => [ws.name, ws.prettyName])),
    [workspaces],
  );

  return (
    <div
      className={`rowInstance-bg-color cl-table flex-grow flex-wrap content-between py-0 overflow-auto scrollbar`}
    >
      <TableInstance
        showGuiIcon={false}
        viewMode={WorkspaceRole.manager}
        extended={true}
        instances={instances}
        workspacePrettyName={workspacePrettyName}
        hasSSHKeys={hasSSHKeys}
        handleManagerSorting={handleManagerSorting}
        showAdvanced={showAdvanced}
        showCheckbox={showCheckbox}
        selectiveDestroy={selectiveDestroy}
        selectToDestroy={selectToDestroy}
      />
    </div>
  );
};

export default TableWorkspace;
