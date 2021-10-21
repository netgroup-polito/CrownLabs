import { FC, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableTemplate from '../TableTemplate/TableTemplate';
import { Template, Workspace } from '../../../utils';
import TableWorkspaceRow from './TableWorkspaceRow';

export interface ITableWorkspaceProps {
  workspaces: Array<Workspace>;
  //viewMode: WorkspaceRole;
}

const TableWorkspace: FC<ITableWorkspaceProps> = ({ ...props }) => {
  const { workspaces /* , viewMode */ } = props;
  const [expandedId, setExpandedId] = useState(['']);

  const handleAccordion = (expanded: boolean, record: Workspace) => {
    setExpandedId([expanded ? record.id : '']);
  };

  const expandRow = (rowId: string) => {
    expandedId[0] === rowId ? setExpandedId(['']) : setExpandedId([rowId]);
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

  return (
    <div
      className={`rowInstance-bg-color cl-table flex-grow flex-wrap content-between py-0 overflow-auto scrollbar`}
    >
      <Table
        rowKey={record => record.id}
        columns={columns}
        size={'middle'}
        dataSource={workspaces}
        pagination={false}
        showHeader={false}
        onExpand={handleAccordion}
        expandable={{
          expandedRowKeys: expandedId,
          // eslint-disable-next-line react/no-multi-comp
          expandIcon: ({ expanded, onExpand, record }) => (
            <CaretRightOutlined
              onClick={e => onExpand(record, e)}
              rotate={expanded ? 90 : 0}
            />
          ),
          // eslint-disable-next-line react/no-multi-comp
          expandedRowRender: record => (
            <TableTemplate
              templates={record.templates!} /* viewMode={viewMode} */
            />
          ),
        }}
      />
    </div>
  );
};

export default TableWorkspace;
