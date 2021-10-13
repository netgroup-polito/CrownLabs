import { FC, useState } from 'react';
import { Space, Tooltip } from 'antd';
import Button from 'antd-button-color';
import { TemplatesTableRowSettings } from '../TemplatesTableRowSettings';
import {
  DesktopOutlined,
  CodeOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import Badge from '../../../common/Badge';
import { Resources, WorkspaceRole } from '../../../../utils';
import { CreateInstanceMutation } from '../../../../generated-types';
import { FetchResult } from 'apollo-link';
import { ModalAlert } from '../../../common/ModalAlert';
import { useInstancesLabelSelectorQuery } from '../../../../generated-types';

export interface ITemplatesTableRowProps {
  id: string;
  name: string;
  gui: boolean;
  persistent: boolean;
  role: WorkspaceRole;
  resources: Resources;
  activeInstances: number;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
  createInstance: (
    id: string
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, any>,
      Record<string, any>
    >
  >;
  expandRow: (activeInstances: number, rowId: string, create: boolean) => void;
}

const infiniteIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="20px">
    <path
      d="M256 256s-48-96-126-96c-54.12 0-98 43-98 96s43.88 96 98 96c30 0 56.45-13.18 78-32m48-64s48 96 126 96c54.12 0 98-43 98-96s-43.88-96-98-96c-29.37 0-56.66 13.75-78 32"
      fill="none"
      stroke="currentColor"
      stroke-linecap="round"
      stroke-miterlimit="10"
      stroke-width="48"
    />
  </svg>
);

const TemplatesTableRow: FC<ITemplatesTableRowProps> = ({ ...props }) => {
  const {
    id,
    name,
    gui,
    persistent,
    role,
    resources,
    activeInstances,
    createInstance,
    editTemplate,
    deleteTemplate,
    expandRow,
  } = props;

  const {
    refetch: refetchInstancesLabelSelector,
  } = useInstancesLabelSelectorQuery({
    variables: { labels: `crownlabs.polito.it/template=${id}` },
  });

  const [showDeleteModal, setShowDeleteModal] = useState(false);
  return (
    <>
      <ModalAlert
        headTitle={name}
        alertMessage="Cannot delete this template"
        alertDescription="A template with active instances cannot be deleted. Please delete al the instances associated with this template."
        alertType="warning"
        buttons={[
          <Button
            shape="round"
            className="ml-2 w-24"
            type="primary"
            onClick={() => setShowDeleteModal(false)}
          >
            Close
          </Button>,
        ]}
        show={showDeleteModal}
        setShow={setShowDeleteModal}
      />
      <div className="w-full flex justify-between py-0">
        <div
          className="flex w-full items-center cursor-pointer"
          onClick={() => expandRow(activeInstances, id, false)}
        >
          <Space size={'middle'}>
            <div className="flex items-center">
              {gui ? (
                <DesktopOutlined
                  style={{ fontSize: '24px', color: '#1c7afd' }}
                />
              ) : (
                <CodeOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
              )}
              <label className="ml-3 cursor-pointer">{name}</label>
              {persistent && (
                <Tooltip
                  title={
                    <>
                      <div className="text-center">
                        These Instances can be stopped and restarted without
                        being deleted.
                      </div>
                      <div className="text-center">
                        Your files won't be deleted in case of an internal
                        misservice of CrownLabs.
                      </div>
                    </>
                  }
                >
                  <div className="text-green-500 ml-3 flex items-center">
                    {infiniteIcon}
                  </div>
                </Tooltip>
              )}
            </div>
          </Space>
        </div>
        <Space size={'small'}>
          <Badge value={activeInstances} size={'small'} />
          <Tooltip
            placement="left"
            title={
              <>
                <div>CPU: {resources.cpu || 'unavailable'} Core</div>
                <div>RAM: {resources.memory || 'unavailable'} GB</div>
                <div>
                  {persistent
                    ? ` DISK: ${resources.disk || 'unavailable'} GB`
                    : ``}
                </div>
              </>
            }
          >
            <Button
              with="link"
              type={'warning'}
              size={'large'}
              className={'px-0'}
            >
              Info
            </Button>
          </Tooltip>
          {role === 'manager' ? (
            <TemplatesTableRowSettings
              id={id}
              editTemplate={editTemplate}
              deleteTemplate={() => {
                refetchInstancesLabelSelector()
                  .then(ils => {
                    if (!ils.data.instanceList?.instances!.length && !ils.error)
                      deleteTemplate(id);
                    else setShowDeleteModal(true);
                  })
                  .catch(err => null);
              }}
            />
          ) : (
            <Tooltip placement="top" title={'Create Instance'}>
              <Button
                onClick={() => {
                  createInstance(id)
                    .then(() => expandRow(1, id, true))
                    .catch(() => null);
                }}
                className="xs:hidden block"
                with="link"
                type="primary"
                size="large"
                icon={<PlayCircleOutlined style={{ fontSize: '22px' }} />}
              />
            </Tooltip>
          )}
          <Button
            onClick={() => {
              createInstance(id)
                .then(() => expandRow(1, id, true))
                .catch(() => null);
            }}
            className="hidden xs:block"
            type="primary"
            shape="round"
            size={'large'}
          >
            Create
          </Button>
        </Space>
      </div>
    </>
  );
};

export default TemplatesTableRow;
