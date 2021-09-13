import { FC, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableTemplate from '../TableTemplate/TableTemplate';
import { Workspace } from '../../../utils';
import TableWorkspaceRow from './TableWorkspaceRow';

export interface ITableWorkspaceProps {
  workspaces: Array<Workspace>;
  //viewMode: WorkspaceRole;
  filter: string;
}

const TableWorkspace: FC<ITableWorkspaceProps> = ({ ...props }) => {
  const { workspaces /* , viewMode */ } = props;

  const columns = [
    {
      title: 'Template',
      key: 'template',
      // eslint-disable-next-line react/no-multi-comp
      render: ({ title, templates }: Workspace) => (
        <TableWorkspaceRow
          text={title}
          nActive={
            templates?.reduce(
              (total, { instances }) => (total += instances.length),
              0
            ) || 0
          }
        />
      ),
    },
  ];

  const [expandedId, setExpandedId] = useState(['']);
  const handleAccordion = (expanded: boolean, record: Workspace) => {
    setExpandedId([expanded ? record.id : '']);
  };
  const data = workspaces;
  return (
    <div
      className={`rowInstance-bg-color cl-table flex-grow flex-wrap content-between py-0 overflow-auto scrollbar`}
    >
      <Table
        rowKey={record => record.id}
        columns={columns}
        size={'middle'}
        dataSource={data}
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
