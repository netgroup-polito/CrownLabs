import {
  CodeOutlined,
  DeleteOutlined,
  DesktopOutlined,
  MoreOutlined,
  AppstoreAddOutlined,
  DockerOutlined,
} from '@ant-design/icons';
import { Badge, Dropdown, Space, Tooltip, Typography } from 'antd';
import { Button } from 'antd';
import { type FC, useMemo, useState } from 'react';
import SvgInfinite from '../../../assets/infinite.svg?react';
import { type Template, WorkspaceRole } from '../../../utils';
import ModalGroupDeletion from '../ModalGroupDeletion/ModalGroupDeletion';

const { Text } = Typography;
export interface ITableTemplateRowProps {
  template: Template;
  destroyAll: () => void;
  expandRow: (rowId: string) => void;
}

const TableTemplateRow: FC<ITableTemplateRowProps> = ({ ...props }) => {
  const { template, destroyAll, expandRow } = props;

  const { id, name, persistent, gui, hasMultipleEnvironments } = template;
  const [showAlert, setShowAlert] = useState(false);

  const { nRunning, nTotal } = useMemo(() => {
    const nTotal = template.instances.length;
    const nRunning = template.instances.filter(i => i.running).length;
    return { nTotal, nRunning };
  }, [template.instances]);

  return (
    <>
      <div
        className="w-full flex justify-between pr-2 cursor-pointer"
        onClick={() => expandRow(id)}
      >
        <Space size="middle">
          {hasMultipleEnvironments ? (
              <Tooltip
                  placement="right"
                  title={
                    <div className="p-2">
                      <div className="font-semibold mb-2 text-center">
                        Multiple Environments ({template.environmentList.length})
                      </div>
                      {template.environmentList.map((env, index) => (
                        <div key={index} className="p-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-medium">{env.name}</span>
                            {env.guiEnabled ? (
                              <div className="flex items-center gap-1.5">
                                <DesktopOutlined style={{ fontSize: '14px', color: '#1c7afd' }} />
                                <span className="text-xs">VM GUI</span>
                                {env.persistent && (
                                  <>
                                    <SvgInfinite width="14px" className="success-color-fg ml-1" />
                                    <span className="text-xs">Persistent</span>
                                  </>
                                )}
                              </div>
                            ) : (
                              env.environmentType === 'Container' ? (
                                <div className="flex items-center gap-1.5">
                                  <DockerOutlined style={{ fontSize: '14px', color: '#1c7afd' }} />
                                  <span className="text-xs">Container SSH</span>
                                  {env.persistent && (
                                    <>
                                      <SvgInfinite width="14px" className="success-color-fg ml-1" />
                                      <span className="text-xs">Persistent</span>
                                    </>
                                  )}
                                </div>
                              ) : (
                                <div className="flex items-center gap-1.5">
                                  <CodeOutlined style={{ fontSize: '14px', color: '#1c7afd' }} />
                                  <span className="text-xs">VM SSH</span>
                                  {env.persistent && (
                                    <>
                                      <SvgInfinite width="14px" className="success-color-fg ml-1" />
                                      <span className="text-xs">Persistent</span>
                                    </>
                                  )}
                                </div>
                              )
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  }
                >
                  <AppstoreAddOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
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
          <Badge
            size="small"
            color="blue"
            count={`${nRunning}/${nTotal}`}
            className="mx-0"
          />
          <Text className="font-bold w-28 xs:w-48 sm:w-max" ellipsis>
            {name}
          </Text>
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
                    disservice of CrownLabs.
                  </div>
                </>
              }
            >
              <div className="success-color-fg flex items-center">
                <SvgInfinite width="22px" />
              </div>
            </Tooltip>
          )}
        </Space>
        <Button
          color="danger"
          variant="text"
          shape="round"
          size="middle"
          icon={<DeleteOutlined className="mr-1" />}
          className="hidden lg:inline-block"
          onClick={e => {
            e.stopPropagation();
            setShowAlert(true);
          }}
        >
          Destroy All
        </Button>
        <Dropdown
          trigger={['click']}
          menu={{
            items: [
              {
                key: 'destroy_all',
                icon: <DeleteOutlined className="text-lg" />,
                danger: true,
              },
            ],
            onClick: () => setShowAlert(true),
          }}
        >
          <Button
            className="lg:hidden flex justify-center"
            color="default"
            type="link"
            shape="circle"
            size="middle"
            onClick={e => e.stopPropagation()}
            icon={
              <MoreOutlined
                className="flex items-center"
                style={{ fontSize: '20px' }}
              />
            }
          />
        </Dropdown>
      </div>
      <ModalGroupDeletion
        view={WorkspaceRole.manager}
        persistent={
          !!template.instances.filter(i => i.persistent === true).length
        }
        groupName={template.name}
        selective={false}
        instanceList={template.instances.map(i => i.id)}
        show={showAlert}
        setShow={setShowAlert}
        destroy={destroyAll}
      />
    </>
  );
};

export default TableTemplateRow;
