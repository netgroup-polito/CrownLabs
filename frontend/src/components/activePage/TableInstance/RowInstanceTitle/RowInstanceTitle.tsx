import { FC, useEffect, useState } from 'react';
import { Typography, Space, Tooltip } from 'antd';
import RowInstanceStatus from '../RowInstanceStatus/RowInstanceStatus';
import { DesktopOutlined, CodeOutlined } from '@ant-design/icons';
import { WorkspaceRole, Instance } from '../../../../utils';
import PersistentIcon from '../../../common/PersistentIcon/PersistentIcon';
import { useApplyInstanceMutation } from '../../../../generated-types';
import { setInstancePrettyname } from '../../../../utilsLogic';

const { Text } = Typography;
export interface IRowInstanceTitleProps {
  viewMode: WorkspaceRole;
  extended: boolean;
  instance: Instance;
  showGuiIcon: boolean;
}

const RowInstanceTitle: FC<IRowInstanceTitleProps> = ({ ...props }) => {
  const { viewMode, extended, instance, showGuiIcon } = props;
  const {
    name,
    prettyName,
    templatePrettyName,
    tenantId,
    tenantDisplayName,
    status,
    persistent,
    gui,
  } = instance;

  const [edit, setEdit] = useState(false);
  const [title, setTitle] = useState(prettyName || name);
  const [invalid, setInvalid] = useState(false);
  const [applyInstanceMutation] = useApplyInstanceMutation();

  const mutateInstancePrettyname = async (title: string) => {
    if (title.length < 5) {
      setInvalid(true);
    } else {
      setTitle(title);
      setInvalid(false);
      try {
        const result = await setInstancePrettyname(
          title,
          instance,
          applyInstanceMutation
        );
        if (result) setTimeout(setEdit, 400, false);
      } catch {
        // TODO: properly handle errors
      }
    }
  };

  const handleEdit = (text: string) => {
    mutateInstancePrettyname(text);
  };

  const cancelEdit = () => {
    setTitle(title);
  };

  useEffect(() => {
    if (prettyName) {
      setTitle(prettyName);
    }
  }, [prettyName]);

  return (
    <>
      <div className="w-full flex justify-start items-center pl-4">
        <Space size={'middle'}>
          <RowInstanceStatus status={status} />

          {viewMode === 'manager' ? (
            <div className="flex items-center gap-4">
              <Text>{tenantId}</Text>
              <Text className="hidden lg:w-32 xl:w-max md:block" ellipsis>
                {tenantDisplayName}
              </Text>
              <Text
                className="hidden md:w-40 xl:w-48 2xl:w-max lg:block"
                ellipsis
              >
                {prettyName ?? name}
              </Text>
            </div>
          ) : (
            <>
              {showGuiIcon && extended && (
                <div className="flex items-center">
                  {gui ? (
                    <DesktopOutlined
                      className="primary-color-fg"
                      style={{ fontSize: '24px' }}
                    />
                  ) : (
                    <CodeOutlined
                      className="primary-color-fg"
                      style={{ fontSize: '24px' }}
                    />
                  )}
                </div>
              )}

              <Tooltip visible={invalid} title="Title must be at least 5 char">
                <Text
                  editable={{
                    tooltip: 'Click to Edit',
                    editing: edit,
                    autoSize: { maxRows: 1 },
                    onChange: value => handleEdit(value),
                    onCancel: cancelEdit,
                  }}
                  className="w-32 lg:w-40 p-0 m-0"
                  onClick={() => setEdit(true)}
                  ellipsis
                >
                  {title}
                </Text>
              </Tooltip>
              {extended && (
                <Text
                  className="md:w-max hidden xs:block xs:w-28 sm:hidden md:block"
                  ellipsis
                >
                  <i>{templatePrettyName}</i>
                </Text>
              )}
              {persistent && extended && <PersistentIcon />}
            </>
          )}
        </Space>
      </div>
    </>
  );
};

export default RowInstanceTitle;
