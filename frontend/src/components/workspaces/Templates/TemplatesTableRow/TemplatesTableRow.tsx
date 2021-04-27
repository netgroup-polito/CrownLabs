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
import { WorkspaceRole } from '../../../../utils';

export interface ITemplatesTableRowProps {
  id: string;
  name: string;
  gui: boolean;
  role: WorkspaceRole;
  activeInstances: number;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}

const TemplatesTableRow: FC<ITemplatesTableRowProps> = ({ ...props }) => {
  const {
    id,
    name,
    gui,
    role,
    activeInstances,
    editTemplate,
    deleteTemplate,
  } = props;

  return (
    <>
      <div className="w-full flex justify-between py-0">
        <Space size={'middle'}>
          {gui ? (
            <DesktopOutlined
              className={'primary-color-fg'}
              style={{ fontSize: '24px' }}
            />
          ) : (
            <CodeOutlined
              className={'primary-color-fg'}
              style={{ fontSize: '24px' }}
            />
          )}
          {name}
        </Space>
        <Space size={'small'}>
          <Badge value={activeInstances} size={'small'} />
          <Tooltip
            placement="top"
            title={'CPU: 2 Core - RAM: 2GB'}
            trigger={'click'}
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
