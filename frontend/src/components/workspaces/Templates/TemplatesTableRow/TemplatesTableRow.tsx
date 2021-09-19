import { FC } from 'react';
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
    editTemplate,
    deleteTemplate,
  } = props;

  return (
    <>
      <div className="w-full flex justify-between py-0">
        <Space size={'middle'}>
          <div className="flex items-center">
            {gui ? (
              <DesktopOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
            ) : (
              <CodeOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
            )}
            <label className="ml-3">{name}</label>
            {persistent && (
              <Tooltip
                title={
                  <>
                    <div className="text-center">
                      These Instances can be stopped and restarted without being
                      deleted.
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
              deleteTemplate={deleteTemplate}
            />
          ) : (
            <Tooltip placement="top" title={'Create Instance'}>
              <Button
                className="xs:hidden block"
                with="link"
                type="primary"
                size="large"
                icon={<PlayCircleOutlined style={{ fontSize: '22px' }} />}
              />
            </Tooltip>
          )}
          <Button
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
