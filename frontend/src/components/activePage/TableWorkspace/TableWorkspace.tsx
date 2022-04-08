import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import {
  Dispatch,
  FC,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { Instance, Template, Workspace } from '../../../utils';
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
    sortingTemplate: string
  ) => void;
  destroySelectedTrigger: boolean;
  setDestroySelectedTrigger: Dispatch<SetStateAction<boolean>>;
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
  } = props;
  const [expandedId, setExpandedId] = useState(expandedWS.get().split(','));
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const expandWorkspace = () => {
    setExpandedId(workspaces.map(ws => ws.id));
  };

  const collapseWorkspace = () => {
    setExpandedId([]);
  };

  const expandRow = (rowId: string) => {
    expandedId.includes(rowId)
      ? setExpandedId(old => old.filter(id => id !== rowId))
      : setExpandedId(old => [...old, rowId]);
  };

  const getActives = (templates?: Template[]) => {
    return (
      templates?.reduce(
        (total, { instances }) => (total += instances.length),
        0
      ) || 0
    );
  };

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
      // eslint-disable-next-line react/no-multi-comp
      render: ({ title, templates, id }: Workspace) => (
        <TableWorkspaceRow
          title={title}
          id={id}
          nActive={getActives(templates)}
          expandRow={expandRow}
        />
      ),
    },
  ];

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
        rowKey={record => record.id}
        columns={columns}
        size="middle"
        dataSource={workspaces}
        pagination={false}
        showHeader={false}
        onExpand={(expanded, ws) => expandRow(ws.id)}
        expandable={{
          expandedRowKeys: expandedId,
          // eslint-disable-next-line react/no-multi-comp
          expandIcon: ({ expanded, onExpand, record }) => (
            <CaretRightOutlined
              className="transition-icon"
              onClick={e => onExpand(record, e)}
              rotate={expanded ? 90 : 0}
            />
          ),
          // eslint-disable-next-line react/no-multi-comp
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
