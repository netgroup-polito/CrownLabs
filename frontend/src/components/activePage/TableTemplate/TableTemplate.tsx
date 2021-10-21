import { FC, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableInstance from '../TableInstance/TableInstance';
import { Instance, Template, WorkspaceRole } from '../../../utils';
import TableTemplateRow from './TableTemplateRow';
import './TableTemplate.less';
import { useDeleteInstanceMutation } from '../../../generated-types';

export interface ITableTemplateProps {
  templates: Array<Template>;
}

const TableTemplate: FC<ITableTemplateProps> = ({ ...props }) => {
  const { templates } = props;

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const [expandedId, setExpandedId] = useState(['']);
  const handleAccordion = (expanded: boolean, record: Template) => {
    setExpandedId(expanded ? [record.id] : ['']);
  };

  const expandRow = (rowId: string) => {
    expandedId[0] === rowId ? setExpandedId(['']) : setExpandedId([rowId]);
  };

  const destroyAll = (templateId: string) => {
    templates
      .find(t => t.id === templateId)!
      .instances.forEach(instance => {
        deleteInstanceMutation({
          variables: {
            tenantNamespace: instance.tenantNamespace!,
            instanceId: instance.name,
          },
        });
      });
  };

  const columns = [
    {
      title: 'Template',
      key: 'template',
      // eslint-disable-next-line react/no-multi-comp
      render: (template: Template) => (
        <TableTemplateRow
          key={template.id}
          template={template}
          nActive={template.instances.length}
          destroyAll={() => destroyAll(template.id)}
          expandRow={expandRow}
        />
      ),
    },
  ];
  return (
    <div
      className={`rowInstance-bg-color ${
        //viewMode === 'user'
        //? 'cl-table flex-grow flex-wrap content-between py-0 overflow-auto scrollbar'
        //:
        ''
      }`}
    >
      <Table
        /* className="rowInstance-bg-color" */
        rowKey={record => record.id}
        columns={columns}
        size={'middle'}
        dataSource={templates}
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
          expandedRowRender: ({ instances }) => (
            <TableInstance
              showGuiIcon={false}
              viewMode={WorkspaceRole.manager}
              extended={true}
              instances={instances as Instance[]}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableTemplate;
