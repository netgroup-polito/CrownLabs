import { FC, useState } from 'react';
import { Table, message } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableInstance from '../TableInstance/TableInstance';
import { Instance, Template, WorkspaceRole } from '../../../utils';
import TableTemplateRow from './TableTemplateRow';
import './TableTemplate.less';
import { useDeleteInstanceMutation } from '../../../generated-types';

export interface ITableTemplateProps {
  templates: Array<Template>;
  //viewMode: WorkspaceRole;
}

const TableTemplate: FC<ITableTemplateProps> = ({ ...props }) => {
  const { templates /* , viewMode */ } = props;

  const columns = [
    {
      title: 'Template',
      key: 'template',
      // eslint-disable-next-line react/no-multi-comp
      render: (record: Template) => (
        <TableTemplateRow
          key={record.id}
          persistent={record.persistent}
          text={record.name}
          nActive={record.instances.length}
          gui={record.gui}
          destroyAll={() => message.info('All VMs deleted')}
        />
      ),
    },
  ];

  const [deleteInstanceMutation] = useDeleteInstanceMutation();
  const [expandedId, setExpandedId] = useState(['']);
  const handleAccordion = (expanded: boolean, record: Template) => {
    expanded ? setExpandedId([record.id]) : setExpandedId(['']);
  };
  const startInstance = (idInstance: string, idTemplate: string) => {};
  const stopInstance = (idInstance: string, idTemplate: string) => {};

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
          expandedRowRender: ({ instances, persistent }) => (
            <TableInstance
              showGuiIcon={false}
              viewMode={WorkspaceRole.manager}
              extended={true}
              instances={instances as Instance[]}
              startInstance={startInstance}
              stopInstance={stopInstance}
              destroyInstance={(instanceId: string, tenantNamespace: string) =>
                deleteInstanceMutation({
                  variables: { tenantNamespace, instanceId },
                })
              }
            />
          ),
        }}
      />
    </div>
  );
};

export default TableTemplate;
