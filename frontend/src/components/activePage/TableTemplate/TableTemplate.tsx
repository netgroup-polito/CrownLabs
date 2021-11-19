import { Dispatch, FC, SetStateAction, useEffect, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import TableInstance from '../TableInstance/TableInstance';
import { Template, User, WorkspaceRole } from '../../../utils';
import TableTemplateRow from './TableTemplateRow';
import './TableTemplate.less';
import {
  useDeleteInstanceMutation,
  useSshKeysQuery,
} from '../../../generated-types';

export interface ITableTemplateProps {
  templates: Array<Template>;
  user: User;
  collapseAll: boolean;
  expandAll: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  showAdvanced: boolean;
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string
  ) => void;
}

const TableTemplate: FC<ITableTemplateProps> = ({ ...props }) => {
  const {
    templates,
    user,
    collapseAll,
    expandAll,
    setCollapseAll,
    setExpandAll,
    handleManagerSorting,
    showAdvanced,
  } = props;
  const { tenantId } = user;
  const [expandedId, setExpandedId] = useState(
    window.sessionStorage
      .getItem(`prevExpandedIdActivePageTemplate-${templates[0].workspaceId}`)
      ?.split(',') ?? []
  );

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const { data: sshKeysResult } = useSshKeysQuery({
    variables: { tenantId: tenantId ?? '' },
    notifyOnNetworkStatusChange: true,
    fetchPolicy: 'network-only',
  });

  const hasSSHKeys = !!sshKeysResult?.tenant?.spec?.publicKeys?.length;

  const expandTemplate = () => {
    setExpandedId(templates.map(t => t.id));
    setExpandAll(false);
  };

  const collapseTemplate = () => {
    setExpandedId([]);
    setCollapseAll(false);
  };

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

  useEffect(() => {
    window.sessionStorage.setItem(
      `prevExpandedIdActivePageTemplate-${templates[0].workspaceId}`,
      expandedId.join(',')
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expandedId]);

  useEffect(() => {
    if (collapseAll) collapseTemplate();
    if (expandAll) expandTemplate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [collapseAll, expandAll]);

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
              instances={instances}
              hasSSHKeys={hasSSHKeys}
              handleManagerSorting={handleManagerSorting}
              showAdvanced={showAdvanced}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableTemplate;
