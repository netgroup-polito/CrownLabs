import { FC } from 'react';
import { Space, Tooltip } from 'antd';
import Button from 'antd-button-color';
import { TemplatesTableRowSettings } from '../TemplatesTableRowSettings';
import {
  DesktopOutlined,
  CodeOutlined,
  PlayCircleOutlined,
  SafetyCertificateOutlined,
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
          {gui ? (
            <DesktopOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
          ) : (
            <CodeOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
          )}
          <div className="flex items-end">
            {name}
            {persistent && (
              <Tooltip title="persistent">
                <SafetyCertificateOutlined
                  className="text-green-500 ml-2 mb-0.5"
                  style={{ fontSize: '18px' }}
                />
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
