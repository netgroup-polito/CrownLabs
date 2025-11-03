import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import type { Dispatch, FC, SetStateAction } from 'react';
import { useContext, useEffect, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { TenantContext } from '../../../contexts/TenantContext';
import type { Template } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import { SessionValue, StorageKeys } from '../../../utilsStorage';
import TableInstance from '../TableInstance/TableInstance';
import './TableTemplate.less';
import TableTemplateRow from './TableTemplateRow';

const expandedT = new SessionValue(StorageKeys.Active_ID_T, '');
export interface ITableTemplateProps {
  templates: Array<Template>;
  collapseAll: boolean;
  expandAll: boolean;
  setCollapseAll: Dispatch<SetStateAction<boolean>>;
  setExpandAll: Dispatch<SetStateAction<boolean>>;
  showAdvanced: boolean;
  showCheckbox: boolean;
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string,
  ) => void;
  selectiveDestroy?: string[];
  selectToDestroy?: (instanceId: string) => void;
}

const TableTemplate: FC<ITableTemplateProps> = ({ ...props }) => {
  const {
    templates,
    collapseAll,
    expandAll,
    setCollapseAll,
    setExpandAll,
    handleManagerSorting,
    showAdvanced,
    showCheckbox,
    selectToDestroy,
    selectiveDestroy,
  } = props;
  const { hasSSHKeys } = useContext(TenantContext);
  const [expandedId, setExpandedId] = useState(
    expandedT.get(templates[0].workspaceName).split(','),
  );
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const expandTemplate = () => {
    setExpandedId(templates.map(t => t.id));
    setExpandAll(false);
  };

  const collapseTemplate = () => {
    setExpandedId([]);
    setCollapseAll(false);
  };

  const expandRow = (rowId: string) =>
    expandedId.includes(rowId)
      ? setExpandedId(old => old.filter(id => id !== rowId))
      : setExpandedId(old => [...old, rowId]);

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
      render: (template: Template) => (
        <TableTemplateRow
          key={template.id}
          template={template}
          destroyAll={() => destroyAll(template.id)}
          expandRow={expandRow}
        />
      ),
    },
  ];

  useEffect(() => {
    expandedT.set(expandedId.join(','), templates[0].workspaceName);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expandedId]);

  useEffect(() => {
    if (collapseAll) collapseTemplate();
    if (expandAll) expandTemplate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [collapseAll, expandAll]);

  return (
    <div className="rowInstance-bg-color">
      <Table
        /* className="rowInstance-bg-color" */
        rowKey={record => record.id}
        columns={columns}
        size="middle"
        dataSource={templates}
        pagination={false}
        showHeader={false}
        expandable={{
          onExpand: (_expanded, ws) => expandRow(ws.id),
          expandedRowKeys: expandedId,
          expandIcon: ({ expanded, onExpand, record }) => (
            <CaretRightOutlined
              className="transition-icon"
              onClick={e => onExpand(record, e)}
              rotate={expanded ? 90 : 0}
            />
          ),
          expandedRowRender: ({ instances }) => (
            <TableInstance
              showGuiIcon={false}
              viewMode={WorkspaceRole.manager}
              extended={true}
              instances={instances}
              hasSSHKeys={hasSSHKeys}
              handleManagerSorting={handleManagerSorting}
              showAdvanced={showAdvanced}
              showCheckbox={showCheckbox}
              selectiveDestroy={selectiveDestroy}
              selectToDestroy={selectToDestroy}
            />
          ),
        }}
      />
    </div>
  );
};

export default TableTemplate;
