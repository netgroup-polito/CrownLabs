import { CaretRightOutlined } from '@ant-design/icons';
import { Table } from 'antd';
import {
  Dispatch,
  FC,
  SetStateAction,
  useContext,
  useEffect,
  useState,
} from 'react';
import { useDeleteInstanceMutation } from '../../../generated-types';
import { TenantContext } from '../../../graphql-components/tenantContext/TenantContext';
import { Template, WorkspaceRole } from '../../../utils';
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
  handleManagerSorting: (
    sortingType: string,
    sorting: number,
    sortingTemplate: string
  ) => void;
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
  } = props;
  const { hasSSHKeys } = useContext(TenantContext);
  const [expandedId, setExpandedId] = useState(
    expandedT.get(templates[0].workspaceId).split(',')
  );

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

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
    expandedT.set(expandedId.join(','), templates[0].workspaceId);
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
        onExpand={(expanded, ws) => expandRow(ws.id)}
        expandable={{
          expandedRowKeys: expandedId,
          // eslint-disable-next-line react/no-multi-comp
          expandIcon: ({ expanded, onExpand, record }) => (
            <CaretRightOutlined
              className="transition-icon"
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
