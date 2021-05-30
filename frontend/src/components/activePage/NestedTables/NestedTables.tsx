/* eslint-disable @typescript-eslint/no-unused-vars */

import { FC, useState } from 'react';
import { Table } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import RowHeading from '../RowHeading/RowHeading';
import InstancesTable from '../Instances/InstancesTable/InstancesTable';
import { IInstance } from '../Instances/InstancesTable/InstancesTable';
import { instances } from '../tempData';

export interface IWorkspace {
  key: string;
  id: string;
  name: string;
  templates: Array<ITemplate>;
}

export interface ITemplate {
  key: string;
  id: string;
  name: string;
  workspace: string;
  nActiveInstances: number;
}
export interface INestedTablesProps {
  workspaces: Array<IWorkspace>;
  templates: Array<ITemplate>;
  isManager: boolean;
  nested: boolean;
  destroyAll: () => void;
}

type rowType = IWorkspace | ITemplate;

const NestedTables: FC<INestedTablesProps> = ({ ...props }) => {
  const { workspaces, templates, nested, isManager, destroyAll } = props;
  const { Column } = Table;
  const [expandedID, setExpandedID] = useState(['']);

  const accordion = (expanded: boolean, record: rowType) => {
    const expId = !expanded
      ? ''
      : nested
      ? (record as IWorkspace).key
      : (record as ITemplate).key;
    setExpandedID([expId]);
  };
  const data = (nested ? workspaces : templates) as rowType[];
  return (
    <Table
      dataSource={data}
      pagination={false}
      showHeader={false}
      expandable={{
        expandedRowKeys: expandedID,
        // eslint-disable-next-line react/no-multi-comp
        expandIcon: ({ expanded, onExpand, record }) => (
          <CaretRightOutlined
            onClick={e => onExpand(record, e)}
            rotate={expanded ? 90 : 0}
          />
        ),
        // eslint-disable-next-line react/no-multi-comp
        expandedRowRender: record =>
          nested ? (
            <NestedTables
              workspaces={workspaces}
              templates={(record as IWorkspace).templates}
              nested={false}
              isManager={isManager}
              destroyAll={destroyAll}
            />
          ) : (
            <InstancesTable
              instances={instances.filter(
                (inst: IInstance) => inst.templateId === record.id
              )}
              isManaged={isManager}
            />
          ),
        expandRowByClick: true,
      }}
      onExpand={accordion}
    >
      <Column
        title="Workspaces"
        dataIndex="name"
        key="key"
        render={(text, record: IWorkspace | ITemplate) => {
          return nested ? (
            <RowHeading
              text={(record as IWorkspace).name}
              nActive={(record as IWorkspace).templates.length}
              newTempl={true}
              destroyAll={destroyAll}
            />
          ) : (
            <RowHeading
              text={(record as ITemplate).name}
              nActive={(record as ITemplate).nActiveInstances}
              newTempl={true}
              destroyAll={destroyAll}
            />
          );
        }}
      />
    </Table>
  );
};

export default NestedTables;
