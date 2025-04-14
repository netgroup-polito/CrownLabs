/* eslint-disable react/no-multi-comp */
import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import { FetchResult } from '@apollo/client';
import { FC, useContext, useEffect, useState } from 'react';
import {
  CreateInstanceMutation,
  DeleteTemplateMutation,
} from '../../../../generated-types';
import { TenantContext } from '../../../../contexts/TenantContext';
import { makeListToggler, Template, WorkspaceRole } from '../../../../utils';
import { SessionValue, StorageKeys } from '../../../../utilsStorage';
import TableInstance from '../../../activePage/TableInstance/TableInstance';
import { TemplatesTableRow } from '../TemplatesTableRow';
import './TemplatesTable.less';

const expandedT = new SessionValue(StorageKeys.Dashboard_ID_T, '');
export interface ITemplatesTableProps {
  totalInstances: number;
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
    id: string,
    labelSelector?: JSON
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
    totalInstances,
    templates,
    role,
    editTemplate,
    deleteTemplate,
    deleteTemplateLoading,
    createInstance,
  } = props;

  const { hasSSHKeys } = useContext(TenantContext);
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
          template={record}
          role={role}
          totalInstances={totalInstances}
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
              hasSSHKeys={hasSSHKeys}
            />
          ),
        }}
      />
    </div>
  );
};

export default TemplatesTable;
