import {
  CodeOutlined,
  DesktopOutlined,
  AppstoreAddOutlined,
  DockerOutlined,
} from '@ant-design/icons';
import { Checkbox, Space, Tooltip, Typography } from 'antd';
import SvgInfinite from '../../../../assets/infinite.svg?react';
import type { ApolloError } from '@apollo/client';
import { type FC, useContext, useEffect, useState } from 'react';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import { useApplyInstanceMutation } from '../../../../generated-types';
import { type Instance, WorkspaceRole } from '../../../../utils';
import { setInstancePrettyname } from '../../../../utilsLogic';
import PersistentIcon from '../../../common/PersistentIcon/PersistentIcon';
import RowInstanceStatus from '../RowInstanceStatus/RowInstanceStatus';
import NodeSelectorIcon from '../../../common/NodeSelectorIcon/NodeSelectorIcon';

const { Text } = Typography;
export interface IRowInstanceTitleProps {
  viewMode: WorkspaceRole;
  extended: boolean;
  instance: Instance;
  showGuiIcon: boolean;
  showCheckbox?: boolean;
  selectiveDestroy?: string[];
  selectToDestroy?: (instanceId: string) => void;
}

const RowInstanceTitle: FC<IRowInstanceTitleProps> = ({ ...props }) => {
  const {
    viewMode,
    extended,
    instance,
    showGuiIcon,
    showCheckbox,
    selectiveDestroy,
    selectToDestroy,
  } = props;
  const {
    name,
    prettyName,
    templatePrettyName,
    tenantId,
    tenantDisplayName,
    status,
    persistent,
    nodeSelector,
    gui,
    hasMultipleEnvironments,
  } = instance;

  const [edit, setEdit] = useState(false);
  const [title, setTitle] = useState(prettyName || name);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const mutateInstancePrettyname = async (title: string) => {
    setTitle(title);
    try {
      const result = await setInstancePrettyname(
        title,
        instance,
        applyInstanceMutation,
      );
      if (result) setTimeout(setEdit, 400, false);
    } catch (err) {
      apolloErrorCatcher(err as ApolloError);
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
      <div className="w-full flex justify-start items-center pl-2">
        <Space size="middle">
          {viewMode === WorkspaceRole.manager &&
            selectiveDestroy &&
            selectToDestroy &&
            showCheckbox && (
              <div className="flex mr-2 items-center">
                <Checkbox
                  checked={selectiveDestroy.includes(instance.id)}
                  className="p-0"
                  onClick={() => selectToDestroy(instance.id)}
                />
              </div>
            )}
          <RowInstanceStatus
            status={status}
            environments={instance.environments}
          />

          {viewMode === WorkspaceRole.manager ? (
            <div className="flex items-center gap-4">
              <Text className="w-32">{tenantId}</Text>
              <Text className="hidden w-max lg:w-32 2xl:w-40 md:block" ellipsis>
                {tenantDisplayName}
              </Text>
              <Text
                className="hidden lg:w-32 xl:w-40 2xl:w-max lg:block"
                ellipsis
              >
                {prettyName ?? name}
              </Text>
            </div>
          ) : (
            <>
              {showGuiIcon && extended && (
                <div className="flex items-center">
                  {instance.environments && hasMultipleEnvironments ? (
                    <Tooltip
                      placement="right"
                      title={
                        <div className="p-2">
                          <div className="font-semibold mb-2 text-center">
                            Multiple Environments (
                            {instance.environments.length})
                          </div>
                          {instance.environments.map((env, index) => (
                            <div key={index} className="p-1">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="font-medium">{env.name}</span>
                                {env.guiEnabled ? (
                                  <div className="flex items-center gap-1.5">
                                    <DesktopOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">VM GUI</span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                ) : env.environmentType === 'Container' ? (
                                  <div className="flex items-center gap-1.5">
                                    <DockerOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">
                                      Container SSH
                                    </span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                ) : (
                                  <div className="flex items-center gap-1.5">
                                    <CodeOutlined
                                      style={{
                                        fontSize: '14px',
                                        color: '#1c7afd',
                                      }}
                                    />
                                    <span className="text-xs">VM SSH</span>
                                    {env.persistent && (
                                      <>
                                        <SvgInfinite
                                          width="14px"
                                          className="success-color-fg ml-1"
                                        />
                                        <span className="text-xs">
                                          Persistent
                                        </span>
                                      </>
                                    )}
                                  </div>
                                )}
                              </div>
                            </div>
                          ))}
                        </div>
                      }
                    >
                      <AppstoreAddOutlined
                        style={{ fontSize: '24px', color: '#1c7afd' }}
                      />
                    </Tooltip>
                  ) : gui ? (
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
              {extended && (
                <Text
                  className="md:w-max hidden xs:block xs:w-28 sm:hidden md:block"
                  ellipsis
                >
                  <i>{templatePrettyName}</i>
                </Text>
              )}
              {persistent && extended && <PersistentIcon />}
              {nodeSelector && extended && (
                <NodeSelectorIcon
                  isOnWorkspace={false}
                  nodeSelector={nodeSelector}
                />
              )}
            </>
          )}
        </Space>
      </div>
    </>
  );
};

export default RowInstanceTitle;
