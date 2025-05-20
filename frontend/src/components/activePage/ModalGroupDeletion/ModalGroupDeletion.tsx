import { DeleteOutlined } from '@ant-design/icons';
import { Button, Divider } from 'antd';
import { type Dispatch, type FC, type SetStateAction, useState } from 'react';
import SvgInfinite from '../../../assets/infinite.svg?react';
import { WorkspaceRole } from '../../../utils';
import { ModalAlert } from '../../common/ModalAlert';

export interface IModalGroupDeletionProps {
  view: WorkspaceRole;
  persistent: boolean;
  selective: boolean;
  groupName?: string;
  instanceList: Array<string>;
  show: boolean;
  setShow: Dispatch<SetStateAction<boolean>>;
  destroy: () => void;
}

const ModalGroupDeletion: FC<IModalGroupDeletionProps> = ({ ...props }) => {
  const {
    view,
    persistent,
    selective,
    groupName,
    instanceList,
    show,
    setShow,
    destroy,
  } = props;
  const [confirmDeletion, setConfirmDeletion] = useState(false);

  const title = selective ? 'Destroy Selected' : 'Destroy All';
  const message = <b>ATTENTION</b>;
  const description = (
    <>
      <div>
        Are you sure that you want to destroy
        {selective ? (
          <>{` the ${instanceList.length} selected instances`}</>
        ) : groupName ? (
          <>
            {' all instances of '}
            <b>
              <i>{groupName}</i>
            </b>
          </>
        ) : (
          ' all instances'
        )}
        ? <br />
        This operation is <u>dangerous and irreversible</u>!
      </div>

      {persistent ? (
        <div className="text-center text-xs">
          <Divider type="horizontal" className="my-3" />
          <div className="flex items-end">
            <i>
              (Seems you are also trying to destroy one or more Persistent
              instances
              <SvgInfinite
                width="16px"
                className="ml-1.5 success-color-fg align-bottom"
              />
              .
              {view === WorkspaceRole.manager
                ? ' You need to confirm their deletion)'
                : ' They will be skipped, you need to MANUALLY destroy them)'}
            </i>
          </div>
        </div>
      ) : (
        ''
      )}
    </>
  );
  const buttons = [
    <Button
      key="destroy_all"
      color="danger"
      shape="round"
      size="middle"
      disabled={!confirmDeletion}
      icon={<DeleteOutlined />}
      className="border-0"
      onClick={() => {
        destroy();
        setShow(false);
      }}
    >
      {title}
    </Button>,
  ];

  const checkbox = {
    confirmCheckbox: confirmDeletion,
    setConfirmCheckbox: setConfirmDeletion,
    checkboxLabel: 'I understand the risk and I want to proceed',
  };

  return (
    <ModalAlert
      headTitle={title}
      show={show}
      message={message}
      description={description}
      type="error"
      buttons={buttons}
      setShow={setShow}
      checkbox={checkbox}
    />
  );
};

export default ModalGroupDeletion;
