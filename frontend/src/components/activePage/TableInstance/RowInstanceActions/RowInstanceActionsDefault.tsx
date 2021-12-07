import { FC, SetStateAction, useState } from 'react';
import { Tooltip } from 'antd';
import Button from 'antd-button-color';
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons';
import { Instance, WorkspaceRole } from '../../../../utils';
import { useDeleteInstanceMutation } from '../../../../generated-types';
import { EnvironmentType } from '../../../../generated-types';
import { ModalAlert } from '../../../common/ModalAlert';
import Text from 'antd/lib/typography/Text';

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
  const {
    prettyName,
    url,
    status,
    gui,
    name,
    tenantNamespace,
    environmentType,
  } = instance;

  const [deleteInstanceMutation] = useDeleteInstanceMutation();

  const titleFromStatus = () => {
    if (!connectDisabled) {
      return gui ? 'Connect to the GUI' : `Connect through SSH`;
    } else
      return (
        <>
          <div>
            <b>{'Impossible to connect:'}</b>
          </div>
          <div>
            {environmentType === EnvironmentType.Container
              ? 'Containers do not support SSH connection yet'
              : `The instance is ${status}`}
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
    status !== 'VmiReady' ||
    (environmentType === EnvironmentType.Container && !gui);

  const font22px = { fontSize: '22px' };

  const connectOptions = gui
    ? { href: url!, target: '_blank' }
    : { onClick: () => setSshModal(true), ghost: true };

  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);

  return (
    <>
      <ModalAlert
        headTitle="Confirm Instance deletion"
        message={
          <Text>
            Do you really want to delete <b>{prettyName}</b> ?
          </Text>
        }
        description={`Instance ID: ${name}`}
        type="warning"
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
            onClick={() =>
              deleteInstanceMutation({
                variables: {
                  instanceId: name,
                  tenantNamespace: tenantNamespace!,
                },
              })
                .then(() => setShowDeleteModalConfirm(false))
                //TODO manage error
                .catch(() => null)
            }
          >
            Delete
          </Button>,
        ]}
        show={showDeleteModalConfirm}
        setShow={setShowDeleteModalConfirm}
      />
      <Tooltip placement="top" title="Destroy">
        <Button
          onClick={() => setShowDeleteModalConfirm(true)}
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
          <Button
            className={`${connectDisabled ? 'pointer-events-none' : ''}`}
            type={classFromProps()}
            shape="round"
            size="middle"
            {...connectOptions}
            disabled={connectDisabled}
          >
            Connect
          </Button>
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
          <Button
            className={`${
              connectDisabled ? 'pointer-events-none' : ''
            } flex items-center justify-center p-0 border-0`}
            with={!extended ? 'link' : undefined}
            type={classFromPropsMobile()}
            shape="circle"
            size="middle"
            {...connectOptions}
            disabled={connectDisabled}
            icon={
              <ExportOutlined
                className="flex items-center justify-center"
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
