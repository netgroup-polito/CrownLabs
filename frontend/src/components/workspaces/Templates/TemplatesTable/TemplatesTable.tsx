import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import type { FetchResult } from '@apollo/client';
import type { FC } from 'react';
import { useContext, useEffect, useState } from 'react';
import type {
  CreateInstanceMutation,
  DeleteTemplateMutation,
} from '../../../../generated-types';
import { TenantContext } from '../../../../contexts/TenantContext';
import type { Template } from '../../../../utils';
import { makeListToggler, WorkspaceRole } from '../../../../utils';
import { SessionValue, StorageKeys } from '../../../../utilsStorage';
import TableInstance from '../../../activePage/TableInstance/TableInstance';
import { TemplatesTableRow } from '../TemplatesTableRow';
import './TemplatesTable.less';

const expandedT = new SessionValue(StorageKeys.Dashboard_ID_T, '');
export interface ITemplatesTableProps {
  totalInstances: number;
  tenantNamespace: string;
  workspaceNamespace: string;
  workspaceName: string;
  templates: Array<Template>;
  role: WorkspaceRole;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
  refreshQuota?: () => void; // Add refresh function
  isPersonal?: boolean;
  editTemplate: (id: string) => void;
  deleteTemplate: (
    id: string,
  ) => Promise<
    FetchResult<
      DeleteTemplateMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  deleteTemplateLoading: boolean;
  createInstance: (
    id: string,
    labelSelector?: JSON,
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
}

const TemplatesTable: FC<ITemplatesTableProps> = ({ ...props }) => {
  const {
    totalInstances,
    tenantNamespace,
    templates,
    role,
    deleteTemplate,
    deleteTemplateLoading,
    createInstance,
    availableQuota,
    refreshQuota,
    isPersonal,
  } = props;

  const { hasSSHKeys } = useContext(TenantContext);
  /**
   * Our Table has just one column which render all rows using a component TemplateTableRow
   */

  const columns = [
    {
      title: 'Template',
      key: 'template',
      render: (record: Template) => (
        <TemplatesTableRow
          template={record}
          role={role}
          totalInstances={totalInstances}
          deleteTemplate={deleteTemplate}
          deleteTemplateLoading={deleteTemplateLoading}
          createInstance={createInstance}
          expandRow={listToggler}
          tenantNamespace={tenantNamespace}
          availableQuota={availableQuota}
          refreshQuota={refreshQuota} // Pass refresh function
          isPersonal={isPersonal}
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
    <Table
      size="middle"
      showHeader={false}
      rowKey={record => record.id}
      columns={columns}
      dataSource={templates}
      pagination={false}
      expandable={{
        onExpand: (_expanded, record) => listToggler(`${record.id}`, false),
        rowExpandable: record => !!record.instances.length,
        expandedRowKeys: expandedId,
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
  );
};

export default TemplatesTable;
