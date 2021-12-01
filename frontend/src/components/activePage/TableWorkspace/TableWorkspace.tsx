import { Dispatch, FC, SetStateAction, useEffect, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableTemplate from '../TableTemplate/TableTemplate';
import { Template, User, Workspace } from '../../../utils';
import TableWorkspaceRow from './TableWorkspaceRow';
import { SessionValue, StorageKeys } from '../../../utilsStorage';

const expandedWS = new SessionValue(StorageKeys.Active_ID_WS, '');
export interface ITableWorkspaceProps {
  workspaces: Array<Workspace>;
  user: User;
  collapseAll: boolean;
  expandAll: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  showAdvanced: boolean;
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string
  ) => void;
}

const TableWorkspace: FC<ITableWorkspaceProps> = ({ ...props }) => {
  const {
    workspaces,
    user,
    collapseAll,
    expandAll,
    setCollapseAll,
    setExpandAll,
    showAdvanced,
    handleManagerSorting,
  } = props;
  const [expandedId, setExpandedId] = useState(expandedWS.get().split(','));

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
              user={user}
              collapseAll={collapseAll}
              expandAll={expandAll}
              setCollapseAll={setCollapseAll}
              setExpandAll={setExpandAll}
              handleManagerSorting={handleManagerSorting}
              showAdvanced={showAdvanced}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableWorkspace;
