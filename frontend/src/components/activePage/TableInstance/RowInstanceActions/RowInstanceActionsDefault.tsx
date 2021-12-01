import { FC, SetStateAction } from 'react';
import { Tooltip, Popconfirm } from 'antd';
import Button from 'antd-button-color';
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons';
import { Instance, WorkspaceRole } from '../../../../utils';
import { useDeleteInstanceMutation } from '../../../../generated-types';
import { EnvironmentType } from '../../../../generated-types';

export interface IRowInstanceActionsDefaultProps {
  extended: boolean;
  instance: Instance;
  viewMode: WorkspaceRole;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
}

const RowInstanceActionsDefault: FC<IRowInstanceActionsDefaultProps> = ({
  ...props
}) => {
  const { extended, instance, viewMode, setSshModal } = props;
  const { url, status, gui, name, tenantNamespace, environmentType } = instance;

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const titleFromStatus = () => {
    if (!connectDisabled) {
      return gui ? 'Connect to this Instance' : `Connect by SSH`;
    } else
      return (
        <>
          <div>
            <b>{'Connection unavailable:'}</b>
          </div>
          <div>
            {environmentType === EnvironmentType.Container
              ? 'Container does not support yet SSH connection'
              : `Status: ${status}`}
          </div>
        </>
      );
  };

  const classFromProps = () => {
    if (!connectDisabled) {
      if (extended) return 'primary';
      else return 'success';
    }
    return 'primary';
  };

  const classFromPropsMobile = () => {
    if (!connectDisabled) {
      if (extended) return 'link';
      else return 'success';
    }
    return 'link';
  };

  const connectDisabled =
    status !== 'VmiReady' || environmentType === EnvironmentType.Container;

  const font22px = { fontSize: '22px' };

  return (
    <>
      <Tooltip placement="top" title="Destroy">
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
            type="link"
            danger
            className={`hidden ${
              extended ? 'sm:block' : 'xs:block'
            } py-0 border-0`}
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
          className={`hidden ${
            extended
              ? viewMode === WorkspaceRole.manager
                ? 'xl:block'
                : 'lg:block'
              : 'sm:block '
          } ${connectDisabled ? 'cursor-not-allowed' : ''}`}
        >
          {gui ? (
            <Button
              className={`${connectDisabled ? 'pointer-events-none' : ''}`}
              type={classFromProps()}
              shape="round"
              size="middle"
              href={url!}
              target="_blank"
              disabled={connectDisabled}
            >
              Connect
            </Button>
          ) : (
            <Button
              className={`${connectDisabled ? 'pointer-events-none' : ''}`}
              onClick={() => setSshModal(true)}
              type={classFromProps()}
              shape="round"
              ghost
              size="middle"
              target="_blank"
              disabled={connectDisabled}
            >
              Connect
            </Button>
          )}
        </div>
        <div
          className={`hidden ${
            extended
              ? `sm:block ${
                  viewMode === WorkspaceRole.manager ? 'xl:hidden' : 'lg:hidden'
                }`
              : 'xs:block sm:hidden'
          } block flex items-center ${
            connectDisabled ? 'cursor-not-allowed' : ''
          }`}
        >
          {gui ? (
            <Button
              className={`${
                connectDisabled ? 'pointer-events-none' : ''
              } flex items-center justify-center p-0 border-0`}
              with={!extended ? 'link' : undefined}
              type={classFromPropsMobile()}
              shape="circle"
              size="middle"
              href={url!}
              target="_blank"
              disabled={connectDisabled}
              icon={
                <ExportOutlined
                  className="flex items-center justify-center"
                  style={font22px}
                />
              }
            />
          ) : (
            <Button
              className={`${
                connectDisabled ? 'pointer-events-none' : ''
              } flex items-center justify-center p-0 border-0`}
              with={!extended ? 'link' : undefined}
              type={classFromPropsMobile()}
              shape="circle"
              size="middle"
              ghost
              href={url!}
              target="_blank"
              disabled={connectDisabled}
              icon={
                <ExportOutlined
                  className="flex items-center justify-center"
                  style={font22px}
                />
              }
            />
          )}
        </div>
      </Tooltip>
    </>
  );
};

export default RowInstanceActionsDefault;
