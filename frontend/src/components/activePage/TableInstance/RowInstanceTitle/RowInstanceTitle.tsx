import { FC, useEffect, useState } from 'react';
import { Typography, Space, Tooltip, Input, Form, Row, Col } from 'antd';
import Button from 'antd-button-color';
import RowInstanceStatus from '../RowInstanceStatus/RowInstanceStatus';
import { DesktopOutlined, CodeOutlined } from '@ant-design/icons';
import { WorkspaceRole, Instance } from '../../../../utils';
import PersistentIcon from '../../../common/PersistentIcon/PersistentIcon';

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

  useEffect(() => {
    setTitle(name);
  }, [name]);

  const [edit, setEdit] = useState(false);
  const [title, setTitle] = useState(prettyName || name);
  const [invalid, setInvalid] = useState(false);

  const checkSet = () => {
    if (title.length < 5) {
      setInvalid(true);
    } else {
      setInvalid(false);
      setEdit(false);
    }
  };

  const [form] = Form.useForm();

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
              <Text className="hidden lg:w-48 2xl:w-max lg:block" ellipsis>
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
              {!edit ? (
                <Tooltip title="Click to Edit">
                  <Text className="w-32" onClick={() => setEdit(true)} ellipsis>
                    {title}
                  </Text>
                </Tooltip>
              ) : (
                <Tooltip
                  visible={invalid}
                  title="Title must be at least 5 char"
                >
                  <Form form={form} onFinish={checkSet}>
                    <Form.Item noStyle>
                      <Row gutter={0}>
                        <Col>
                          <Form.Item noStyle rules={[{ required: true }]}>
                            <Input
                              size="small"
                              value={title}
                              onChange={event => setTitle(event.target.value)}
                              style={{
                                width: `${22 + title.length * 8}px`,
                              }}
                              minLength={0}
                            />
                          </Form.Item>
                        </Col>
                        <Col>
                          <Button
                            size={'small'}
                            type="primary"
                            htmlType="submit"
                          >
                            OK
                          </Button>
                        </Col>
                      </Row>
                    </Form.Item>
                  </Form>
                </Tooltip>
              )}
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
