import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import type { Dispatch, FC, SetStateAction } from 'react';
import { useContext, useEffect, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import type { Instance, Workspace } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import TableTemplate from '../TableTemplate/TableTemplate';
import TableWorkspaceRow from './TableWorkspaceRow';

const expandedWS = new SessionValue(StorageKeys.Active_ID_WS, '');
export interface ITableWorkspaceProps {
  instances: Array<Instance>;
  workspaces: Array<Workspace>;
  collapseAll: boolean;
  expandAll: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
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
    collapseAll,
    expandAll,
    setCollapseAll,
    setExpandAll,
    showAdvanced,
    showCheckbox,
    handleManagerSorting,
    destroySelectedTrigger,
    setDestroySelectedTrigger,
    selectiveDestroy,
    selectToDestroy,
    setSelectedPersistent,
  } = props;
  const [expandedId, setExpandedId] = useState(expandedWS.get().split(','));
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const expandWorkspace = () => {
    setExpandedId(workspaces.map(ws => ws.name));
  };

  const collapseWorkspace = () => {
    setExpandedId([]);
  };

  const expandRow = (rowId: string) =>
    expandedId.includes(rowId)
      ? setExpandedId(old => old.filter(id => id !== rowId))
      : setExpandedId(old => [...old, rowId]);

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

  const columns = [
    {
      title: 'Template',
      key: 'template',
      render: ({ prettyName, templates, name }: Workspace) => (
        <TableWorkspaceRow
          title={prettyName}
          id={name}
          templates={templates || []}
          expandRow={expandRow}
        />
      ),
    },
  ];

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
    expandedWS.set(expandedId.join(','));
  }, [expandedId]);

  useEffect(() => {
    if (collapseAll) collapseWorkspace();
    if (expandAll) expandWorkspace();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [collapseAll, expandAll]);

  useEffect(() => {
    if (destroySelectedTrigger) {
      setDestroySelectedTrigger(false);
      destroySelected();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [destroySelectedTrigger]);

  return (
    <div
      className={`rowInstance-bg-color cl-table flex-grow flex-wrap content-between py-0 overflow-auto scrollbar`}
    >
      <Table
        rowKey={record => record.name}
        columns={columns}
        size="middle"
        dataSource={workspaces}
        pagination={false}
        showHeader={false}
        expandable={{
          onExpand: (_expanded, ws) => expandRow(ws.name),
          expandedRowKeys: expandedId,
          expandIcon: ({ expanded, onExpand, record }) => (
            <CaretRightOutlined
              className="transition-icon"
              onClick={e => onExpand(record, e)}
              rotate={expanded ? 90 : 0}
            />
          ),
          expandedRowRender: record => (
            <TableTemplate
              templates={record.templates!}
              collapseAll={collapseAll}
              expandAll={expandAll}
              setCollapseAll={setCollapseAll}
              setExpandAll={setExpandAll}
              handleManagerSorting={handleManagerSorting}
              showAdvanced={showAdvanced}
              showCheckbox={showCheckbox}
              selectiveDestroy={selectiveDestroy}
              selectToDestroy={selectToDestroy}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableWorkspace;
