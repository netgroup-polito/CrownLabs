/* eslint-disable react/no-multi-comp */
import { FC, useEffect } from 'react';
import { Table } from 'antd';
import { TemplatesTableRow } from '../TemplatesTableRow';
import { CaretRightOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { makeListToggler, Template, WorkspaceRole } from '../../../../utils';
import './TemplatesTable.less';
import {
  CreateInstanceMutation,
  DeleteTemplateMutation,
} from '../../../../generated-types';
import { FetchResult } from 'apollo-link';
import TableInstance from '../../../activePage/TableInstance/TableInstance';
import { SessionValue, StorageKeys } from '../../../../utilsStorage';

const expandedT = new SessionValue(StorageKeys.Dashboard_ID_T, '');
export interface ITemplatesTableProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  templates: Array<Template>;
  role: WorkspaceRole;
  editTemplate: (id: string) => void;
  deleteTemplate: (
    id: string
  ) => Promise<
    FetchResult<
      DeleteTemplateMutation,
      Record<string, any>,
      Record<string, any>
    >
  >;
  deleteTemplateLoading: boolean;
  createInstance: (
    id: string
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, any>,
      Record<string, any>
    >
  >;
}

const TemplatesTable: FC<ITemplatesTableProps> = ({ ...props }) => {
  const {
    templates,
    role,
    editTemplate,
    deleteTemplate,
    deleteTemplateLoading,
    createInstance,
  } = props;

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
          deleteTemplateLoading={deleteTemplateLoading}
          createInstance={createInstance}
          expandRow={listToggler}
        />
      ),
    },
  ];

  const [expandedId, setExpandedId] = useState(expandedT.get().split(','));

  const listToggler = makeListToggler<string>(setExpandedId);

  useEffect(() => {
    expandedT.set(expandedId.join(','));
  }, [expandedId]);

  return (
    <div className="w-full flex-grow flex-wrap content-between py-0 overflow-auto scrollbar cl-templates-table">
      <Table
        size="middle"
        showHeader={false}
        rowKey={record => record.id}
        columns={columns}
        dataSource={templates}
        pagination={false}
        onExpand={(expanded, record) => listToggler(`${record.id}`, false)}
        expandable={{
          rowExpandable: record => !!record.instances.length,
          expandedRowKeys: expandedId,
          // eslint-disable-next-line react/no-multi-comp
          expandIcon: ({ expanded, onExpand, record }) =>
            record.instances.length ? (
              <CaretRightOutlined
                className="transition-icon"
                onClick={e => onExpand(record, e)}
                rotate={expanded ? 90 : 0}
              />
            ) : (
              false
            ),
          /**
           * Here we render the expandable content, for example with a nested Table
           */
          // eslint-disable-next-line react/no-multi-comp
          expandedRowRender: template => (
            <TableInstance
              showGuiIcon={false}
              viewMode={WorkspaceRole.user}
              extended={false}
              instances={template.instances}
            />
          ),
        }}
      />
    </div>
  );
};

export default TemplatesTable;
