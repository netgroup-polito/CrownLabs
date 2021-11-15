import { FC } from 'react';
import { Tooltip, Popconfirm } from 'antd';
import Button from 'antd-button-color';
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons';
import { Instance, WorkspaceRole } from '../../../../utils';
import { useDeleteInstanceMutation } from '../../../../generated-types';

export interface IRowInstanceActionsDefaultProps {
  extended: boolean;
  instance: Instance;
  viewMode: WorkspaceRole;
}

const RowInstanceActionsDefault: FC<IRowInstanceActionsDefaultProps> = ({
  ...props
}) => {
  const { extended, instance, viewMode } = props;
  const { url, status, gui, name, tenantNamespace } = instance;

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const titleFromStatus = () => {
    if (status === 'VmiReady') {
      return gui
        ? 'Connect to this Instance'
        : `This instance hasn't any GUI to connect`;
    } else return `Connection unavailable - Status: ` + status;
  };

  const classFromProps = () => {
    if (extended) {
      return 'primary';
    } else {
      if (status === 'VmiReady' && gui) {
        return 'success';
      } else return 'primary';
    }
  };

  const font22px = { fontSize: '22px' };

  return (
    <>
      <Tooltip placement="top" title={'Destroy'}>
        <Popconfirm
          placement="left"
          title="Are you sure to delete?"
          okText="Yes"
          cancelText="No"
          onConfirm={() =>
            deleteInstanceMutation({
              variables: {
                instanceId: name,
                tenantNamespace: tenantNamespace!,
              },
            })
          }
          onCancel={e => e?.stopPropagation()}
        >
          <Button
            className={`hidden ${extended ? 'sm:block' : 'xs:block'} py-0`}
            type="danger"
            with="link"
            shape="circle"
            size="middle"
            icon={
              <DeleteOutlined
                className="flex justify-center items-center"
                style={font22px}
              />
            }
          />
        </Popconfirm>
      </Tooltip>
      <Tooltip placement="top" title={titleFromStatus()}>
        <div
          className={` hidden ${
            extended
              ? viewMode === WorkspaceRole.manager
                ? 'xl:block'
                : 'lg:block'
              : 'sm:block '
          } ${status !== 'VmiReady' || !gui ? 'cursor-not-allowed' : ''}`}
        >
          <Button
            className={`${
              status !== 'VmiReady' || !gui ? 'pointer-events-none' : ''
            }`}
            type={classFromProps()}
            shape="round"
            size={'middle'}
            onClick={() => window.open(url!, '_blank')}
            disabled={status !== 'VmiReady' || !gui}
          >
            Connect
          </Button>
        </div>
      </Tooltip>
      <Tooltip placement="top" title={titleFromStatus()}>
        <div
          className={` hidden ${
            extended
              ? `sm:block ${
                  viewMode === WorkspaceRole.manager ? 'xl:hidden' : 'lg:hidden'
                }`
              : 'xs:block sm:hidden'
          } block flex items-center ${
            status !== 'VmiReady' || !gui ? 'cursor-not-allowed' : ''
          }`}
        >
          <Button
            className={`${
              status !== 'VmiReady' || !gui ? 'pointer-events-none' : ''
            }`}
            with="link"
            type={classFromProps()}
            shape="circle"
            size="middle"
            onClick={() => window.open(url!, '_blank')}
            disabled={status !== 'VmiReady' || !gui}
            icon={
              <ExportOutlined
                className="flex justify-center items-center"
                style={font22px}
              />
            }
          />
        </div>
      </Tooltip>
    </>
  );
};

export default RowInstanceActionsDefault;
