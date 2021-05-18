import { FC } from 'react';
import { Col, Table } from 'antd';
import { TemplatesTableRow } from '../TemplatesTableRow';
import { DownOutlined, RightOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { WorkspaceRole } from '../../../../utils';

export interface ITemplatesTableProps {
  templates: Array<{
    id: string;
    name: string;
    gui: boolean;
    instances: Array<{
      id: number;
      name: string;
      ip: string;
      status: boolean;
    }>;
  }>;
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
      /* eslint-disable react/no-multi-comp */
      render: (record: {
        id: string;
        name: string;
        gui: boolean;
        instances: Array<{
          id: number;
          name: string;
          ip: string;
          status: boolean;
        }>;
      }) => (
        <TemplatesTableRow
          id={record.id}
          name={record.name}
          gui={record.gui}
          role={role}
          activeInstances={record.instances ? record.instances.length : 0}
          editTemplate={editTemplate}
          deleteTemplate={deleteTemplate}
        />
      ),
      /* eslint-enable react/no-multi-comp */
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

  const handleAccordion = (
    expanded: boolean,
    record: {
      id: string;
      name: string;
      gui: boolean;
      instances: Array<{
        id: number;
        name: string;
        ip: string;
        status: boolean;
      }>;
    }
  ) => {
    expanded ? setExpandedId([record.id]) : setExpandedId(['']);
  };

  return (
    <div
      className="w-full flex-grow flex-wrap content-between py-0 overflow-auto scrollbar"
      style={{ height: 'calc(100vh - 380px)', minHeight: '171px' }} //465px if we want to add a Box footer
    >
      <Col span={24} className="flex-auto">
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
            expandIcon: ({ expanded, onExpand, record }) =>
              record.instances.length ? (
                expanded ? (
                  <DownOutlined onClick={e => onExpand(record, e)} />
                ) : (
                  <RightOutlined onClick={e => onExpand(record, e)} />
                )
              ) : (
                false
              ),
            /**
             * Here we render the expandable content, for example with a nested Table
             */
            expandedRowRender: template => 'Running Instances',
          }}
        />
      </Col>
    </div>
  );
};

export default TemplatesTable;
