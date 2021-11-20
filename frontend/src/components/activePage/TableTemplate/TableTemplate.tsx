import { FC, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableInstance from '../TableInstance/TableInstance';
import { Instance, Template, User, WorkspaceRole } from '../../../utils';
import TableTemplateRow from './TableTemplateRow';
import './TableTemplate.less';
import {
  useDeleteInstanceMutation,
  useSshKeysQuery,
} from '../../../generated-types';

export interface ITableTemplateProps {
  templates: Array<Template>;
  user: User;
}

const TableTemplate: FC<ITableTemplateProps> = ({ ...props }) => {
  const { templates, user } = props;
  const { tenantId } = user;
  const [expandedId, setExpandedId] = useState(['']);

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const { data: sshKeysResult } = useSshKeysQuery({
    variables: { tenantId: tenantId ?? '' },
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  });

  const hasSSHKeys = !!sshKeysResult?.tenant?.spec?.publicKeys?.length;

  const expandRow = (rowId: string) => {
    expandedId.includes(rowId)
      ? setExpandedId(old => old.filter(id => id !== rowId))
      : setExpandedId(old => [...old, rowId]);
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
        onExpand={(expanded, ws) => expandRow(ws.id)}
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
              hasSSHKeys={hasSSHKeys}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableTemplate;
