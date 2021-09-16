/* eslint-disable react/no-multi-comp */
import { FC } from 'react';
import { Table } from 'antd';
import { TemplatesTableRow } from '../TemplatesTableRow';
import { RightOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { Template, WorkspaceRole } from '../../../../utils';
import './TemplatesTable.less';

export interface ITemplatesTableProps {
  templates: Array<Template>;
  role: WorkspaceRole;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}

const TemplatesTable: FC<ITemplatesTableProps> = ({ ...props }) => {
  const { templates, role, editTemplate, deleteTemplate } = props;

  /**
   * Our Table has just one column which render all rows using a component TemplateTableRow
   */

  const columns = [
    {
      title: 'Template',
      key: 'template',
      // eslint-disable-next-line react/no-multi-comp
      render: (record: Template) => (
        <TemplatesTableRow
          id={record.id}
          name={record.name}
          gui={record.gui}
          persistent={record.persistent}
          resources={record.resources}
          role={role}
          activeInstances={record.instances ? record.instances.length : 0}
          editTemplate={editTemplate}
          deleteTemplate={deleteTemplate}
        />
      ),
    },
  ];

  /**
   * Handle to manage accordion effect, it's possible to allow multiple row expansion (no accordion effect) by switching
   * handleAccordion() function with the following code
    { expanded
        ? setExpanedId(expandedId => {
            expandedId.push(record.id);
            return expandedId;
          })
        : setExpanedId(expandedId => {
            return expandedId.filter(id => id !== record.id);
          });
    };
   */
  const [expandedId, setExpandedId] = useState(['']);

  const handleAccordion = (expanded: boolean, record: Template) => {
    expanded ? setExpandedId([record.id]) : setExpandedId(['']);
  };

  return (
    <div className="w-full flex-grow flex-wrap content-between py-0 overflow-auto scrollbar cl-templates-table">
      <Table
        size={'small'}
        showHeader={false}
        rowKey={record => record.id}
        columns={columns}
        dataSource={templates}
        pagination={false}
        onExpand={handleAccordion}
        expandable={{
          rowExpandable: record => !!record.instances.length,
          expandedRowKeys: expandedId,
          // eslint-disable-next-line react/no-multi-comp
          expandIcon: ({ expanded, onExpand, record }) =>
            record.instances.length ? (
              <RightOutlined
                onClick={e => onExpand(record, e)}
                rotate={expanded ? 90 : 0}
              />
            ) : (
              false
            ),
          /**
           * Here we render the expandable content, for example with a nested Table
           */
          expandedRowRender: template => 'Instances',
        }}
      />
    </div>
  );
};

export default TemplatesTable;
