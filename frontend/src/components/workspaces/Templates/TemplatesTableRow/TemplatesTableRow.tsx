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
import {
  CreateInstanceMutation,
  DeleteTemplateMutation,
} from '../../../../generated-types';
import { FetchResult } from 'apollo-link';
import { ModalAlert } from '../../../common/ModalAlert';
import { useInstancesLabelSelectorQuery } from '../../../../generated-types';
import { ReactComponent as SvgInfinite } from '../../../../assets/infinite.svg';

export interface ITemplatesTableRowProps {
  id: string;
  name: string;
  gui: boolean;
  persistent: boolean;
  role: WorkspaceRole;
  resources: Resources;
  activeInstances: number;
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
  expandRow: (activeInstances: number, rowId: string, create: boolean) => void;
}

const convertInG = (s: string) =>
  s.includes('M') && Number(s.split('M')[0]) >= 1000
    ? `${Number(s.split('M')[0]) / 1000}G`
    : s;

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
    deleteTemplateLoading,
    expandRow,
  } = props;

  const { refetch: refetchInstancesLabelSelector } =
    useInstancesLabelSelectorQuery({
      variables: { labels: `crownlabs.polito.it/template=${id}` },
    });

  const [showDeleteModalNotPossible, setShowDeleteModalNotPossible] =
    useState(false);
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);
  const [createDisabled, setCreateDisabled] = useState(false);

  const createInstanceHandler = () => {
    setCreateDisabled(true);
    createInstance(id)
      .then(() => {
        setTimeout(() => {
          setCreateDisabled(false);
        }, 400);
        expandRow(1, id, true);
      })
      .catch(() => setCreateDisabled(false));
  };

  return (
    <>
      <ModalAlert
        headTitle={name}
        alertMessage="Cannot delete this template"
        alertDescription="A template with active instances cannot be deleted. Please delete al the instances associated with this template."
        alertType="warning"
        buttons={[
          <Button
            key={0}
            shape="round"
            className="w-24"
            type="primary"
            onClick={() => setShowDeleteModalNotPossible(false)}
          >
            Close
          </Button>,
        ]}
        show={showDeleteModalNotPossible}
        setShow={setShowDeleteModalNotPossible}
      />
      <ModalAlert
        headTitle={name}
        alertMessage="Delete template"
        alertDescription="Do you really want to delete this template?"
        alertType="warning"
        buttons={[
          <Button
            key={0}
            shape="round"
            className="mr-2 w-24"
            type="primary"
            onClick={() => setShowDeleteModalConfirm(false)}
          >
            Close
          </Button>,
          <Button
            key={1}
            shape="round"
            className="ml-2 w-24"
            type="danger"
            loading={deleteTemplateLoading}
            onClick={() =>
              deleteTemplate(id)
                .then(() => setShowDeleteModalConfirm(false))
                .catch(err => null)
            }
          >
            {!deleteTemplateLoading && 'Delete'}
          </Button>,
        ]}
        show={showDeleteModalConfirm}
        setShow={setShowDeleteModalConfirm}
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
                    <SvgInfinite width="22px" />
                  </div>
                </Tooltip>
              )}
            </div>
          </Space>
        </div>
        <Space size={'small'}>
          <Badge value={activeInstances} size={'small'} className="mx-2" />
          <Tooltip
            placement="left"
            title={
              <>
                <div>CPU: {resources.cpu || 'unavailable'} Core</div>
                <div>RAM: {convertInG(resources.memory) || 'unavailable'}B</div>
                <div>
                  {persistent
                    ? ` DISK: ${convertInG(resources.disk) || 'unavailable'}B`
                    : ``}
                </div>
              </>
            }
          >
            <Button
              with="link"
              type={'warning'}
              size={'middle'}
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
                      setShowDeleteModalConfirm(true);
                    else setShowDeleteModalNotPossible(true);
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
            onClick={createInstanceHandler}
            className="hidden xs:block"
            disabled={createDisabled}
            type="primary"
            shape="round"
            size={'middle'}
          >
            Create
          </Button>
        </Space>
      </div>
    </>
  );
};

export default TemplatesTableRow;
