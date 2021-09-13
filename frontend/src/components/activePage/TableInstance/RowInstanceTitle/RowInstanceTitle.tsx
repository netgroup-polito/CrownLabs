import { FC, useState } from 'react';
import { Typography, Space, Tooltip, Input, Form, Row, Col } from 'antd';
import Button from 'antd-button-color';
import RowInstanceStatus from '../RowInstanceStatus/RowInstanceStatus';
import { DesktopOutlined, CodeOutlined } from '@ant-design/icons';
import { SafetyCertificateOutlined } from '@ant-design/icons';
import { WorkspaceRole, Instance } from '../../../../utils';

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
    idTemplate,
    tenantId,
    tenantDisplayName,
    status,
    persistent,
    gui,
  } = instance;

  const [edit, setEdit] = useState(false);
  const [title, setTitle] = useState(name);
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

  return (
    <>
      <div className="w-full flex justify-start items-center pl-4">
        <Space size={'middle'}>
          <RowInstanceStatus status={status} />

          {viewMode === 'manager' ? (
            <div className="flex items-center gap-4">
              <Text>{tenantId}</Text>
              <Text className="hidden md:block">{tenantDisplayName}</Text>
              <Text className="hidden lg:block">{name}</Text>
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
                  <Text onClick={() => setEdit(true)}>{title}</Text>
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
                          <Form.Item
                            noStyle
                            rules={[
                              {
                                required: true,
                              },
                            ]}
                          >
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

                  {/* <Input
                    size="small"
                    value={title}
                    onChange={event => setTitle(event.target.value)}
                    style={{ width: `${18 + title.length * 7}px` }}
                    minLength={10}
                    onPressEnter={checkSet}
                    suffix={
                  <EditOutlined
                    onClick={() => null}
                    className="primary-color-fg flex items-center"
                  />
                }
                  /> */}
                </Tooltip>
              )}
              <Text>
                <i>{idTemplate}</i>
              </Text>
              {persistent && extended && (
                <Tooltip title="Persistent">
                  <SafetyCertificateOutlined
                    onClick={() => null}
                    className="text-green-500 flex items-center"
                    style={{ fontSize: '18px' }}
                  />
                </Tooltip>
              )}
            </>
          )}
        </Space>
      </div>
    </>
  );
};

export default RowInstanceTitle;
