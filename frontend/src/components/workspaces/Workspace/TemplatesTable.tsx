import { FC } from 'react';
import { Col, Table } from 'antd';
import { TemplatesTableRow, InstancesTable } from './';

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
  createInstance: (id: string) => void;
  destroyInstance: (idInstance: number, idTemplate: string) => void;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}

const TemplatesTable: FC<ITemplatesTableProps> = ({ ...props }) => {
  const {
    templates,
    createInstance,
    destroyInstance,
    editTemplate,
    deleteTemplate,
  } = props;

  const templatesRows = templates.map((template, index) =>
    Object.assign({}, template, {
      key: template.id,
      template: (
        <TemplatesTableRow
          id={template.id}
          name={template.name}
          gui={template.gui}
          activeInstances={template.instances ? template.instances.length : 0}
          createInstance={createInstance}
          editTemplate={editTemplate}
          deleteTemplate={deleteTemplate}
        />
      ),
    })
  );

  const instancesTable = (
    id: string,
    instances: Array<{ id: number; name: string; ip: string; status: boolean }>,
    destroyInstance: (idInstance: number, idTemplate: string) => void
  ) => (
    <InstancesTable
      id={id}
      instances={instances}
      destroyInstance={destroyInstance}
    />
  );

  const columns = [
    //E' una sola colonna
    {
      title: 'Template',
      dataIndex: 'template',
      key: 'template',
    },
  ];

  return (
    <div
      className="w-full flex-grow flex-wrap content-between py-0 overflow-auto scrollbar"
      style={{ height: 'calc(100vh - 380px)' }} //465px with footer
    >
      <Col span={24} className="flex-auto">
        <Table
          size={'small'}
          showHeader={false}
          columns={columns} //colonne tabella esterna (templates)
          dataSource={templatesRows} //array di templates (righe già renderizzate), ogni template ha al suo interno un array di instances
          pagination={false}
          expandable={{
            //per ogni row
            //contenuto dopo l'espansione di una row
            expandedRowRender: template =>
              instancesTable(template.id, template.instances, destroyInstance), //renderizzo una table di instances grazie alla funzione instancesRow(...)
            // condizione affinchè una row sia espandibile
            rowExpandable: template =>
              template.instances.length ? true : false,
          }}
        />
      </Col>
    </div>
  );
};

export default TemplatesTable;
