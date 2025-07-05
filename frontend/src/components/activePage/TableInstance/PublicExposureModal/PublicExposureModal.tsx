import {
  Modal,
  Button,
  Form,
  Input,
  Row,
  Col,
  Divider,
  type FormRule,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';

import { type FC, useCallback } from 'react';

interface IPublicExposureModalProps {
  open: boolean;
  onCancel: () => void;
}
interface PortField {
  name?: string;
  targetPort: string;
  desiredPort?: string;
}
interface FormValues {
  ports: PortField[];
}

export const PublicExposureModal: FC<IPublicExposureModalProps> = ({
  open,
  onCancel,
}) => {
  const [form] = Form.useForm<FormValues>();

  const ports = Form.useWatch('ports', form);
  const lastTargetPort = ports?.[ports.length - 1]?.targetPort;
  const isAddDisabled =
    !lastTargetPort ||
    !/^\d+$/.test(lastTargetPort) ||
    parseInt(lastTargetPort, 10) === 0;

  const onFinish = (values: FormValues) => {
    const normalized = values.ports.map(p => ({
      ...p,
      desiredPort: p.desiredPort || '0',
    }));
    console.log('Ports submitted:', normalized);
    onCancel();
  };

  const portValidator = useCallback((_rule: FormRule, value: string) => {
    if (!value) {
      return Promise.resolve();
    }
    const num = parseInt(value, 10);
    if (isNaN(num) || num <= 0 || num > 65535) {
      return Promise.reject(new Error('Port must be between 1 and 65535'));
    }
    return Promise.resolve();
  }, []);

  return (
    <Modal
      title="Port Exposure"
      open={open}
      onCancel={onCancel}
      width={550}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          Close
        </Button>,
        <Button key="send" type="primary" onClick={() => form.submit()}>
          Send
        </Button>,
      ]}
    >
      <Form
        form={form}
        name="dynamic_port_form"
        onFinish={onFinish}
        autoComplete="off"
        layout="vertical"
        initialValues={{
          ports: [{ name: '', targetPort: '', desiredPort: '' }],
        }}
      >
        <Form.List name="ports">
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name, ...restField }, index) => (
                <div key={key}>
                  <Row gutter={8} align="bottom">
                    <Col span={7}>
                      <Form.Item
                        {...restField}
                        name={[name, 'name']}
                        label="Name"
                        style={{ marginBottom: 8 }}
                      >
                        <Input placeholder="e.g. web-server" />
                      </Form.Item>
                    </Col>
                    <Col span={7}>
                      <Form.Item
                        {...restField}
                        name={[name, 'targetPort']}
                        label="Target Port"
                        style={{ marginBottom: 8 }}
                        rules={[
                          { required: true, message: 'Required' },
                          { validator: portValidator },
                        ]}
                      >
                        <Input placeholder="e.g. 8080" />
                      </Form.Item>
                    </Col>
                    <Col span={7}>
                      <Form.Item
                        {...restField}
                        name={[name, 'desiredPort']}
                        label="Desired Port"
                        style={{ marginBottom: 8 }}
                        rules={[{ validator: portValidator }]}
                      >
                        <Input placeholder="Defaults to a random port" />
                      </Form.Item>
                    </Col>
                    <Col
                      span={3}
                      style={{ textAlign: 'center', paddingBottom: '8px' }}
                    >
                      <Button
                        type="text"
                        danger
                        icon={<DeleteOutlined />}
                        onClick={() => remove(name)}
                        disabled={fields.length === 1}
                      />
                    </Col>
                  </Row>
                  {index < fields.length - 1 && (
                    <Divider style={{ margin: '12px 0' }} />
                  )}
                </div>
              ))}
              <Form.Item style={{ textAlign: 'center', marginTop: 24 }}>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  disabled={isAddDisabled}
                >
                  + Add Port
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Form>
    </Modal>
  );
};
