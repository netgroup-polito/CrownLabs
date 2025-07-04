import { Modal, Button, Form, Input, Row, Col, Divider } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { type FC } from 'react';

interface IPublicExposureModalProps {
  open: boolean;
  onCancel: () => void;
}

interface PortField {
  key: number;
  name?: string;
  targetPort: string;
  desiredPort: string;
}

export const PublicExposureModal: FC<IPublicExposureModalProps> = ({
  open,
  onCancel,
}) => {
  const [fields, setFields] = useState<PortField[]>([
    { key: Date.now(), name: '', targetPort: '', desiredPort: '' },
  ]);
  const [errorKeys, setErrorKeys] = useState<number[]>([]);

  const isValidPort = (val: string) => val !== '' && val !== '0';

  const lastField = fields[fields.length - 1];
  const isAddDisabled = !isValidPort(lastField.targetPort);

  const addField = () => {
    if (!isValidPort(lastField.targetPort)) {
      setErrorKeys([lastField.key]);
      return;
    }
    setErrorKeys([]);
    setFields([
      ...fields,
      { key: Date.now(), name: '', targetPort: '', desiredPort: '' },
    ]);
  };

  const removeField = (keyToRemove: number) => {
    if (fields.length > 1) {
      setFields(fields.filter(f => f.key !== keyToRemove));
    }
  };

  const updateField = (
    key: number,
    name: 'targetPort' | 'desiredPort' | 'name',
    value: string,
  ) => {
    if (
      name === 'targetPort' &&
      errorKeys.includes(key) &&
      isValidPort(value)
    ) {
      setErrorKeys(errorKeys.filter(k => k !== key));
    }
    setFields(fields.map(f => (f.key === key ? { ...f, [name]: value } : f)));
  };

  const handleSend = () => {
    const invalid = fields
      .filter(f => !isValidPort(f.targetPort))
      .map(f => f.key);
    if (invalid.length) {
      setErrorKeys(invalid);
      return;
    }
    const normalized = fields.map(f => ({
      ...f,
      desiredPort: f.desiredPort || '0',
    }));
    console.log('Ports:', normalized);
    setErrorKeys([]);
    onCancel();
  };

  return (
    <Modal
      title="Port Exposure"
      open={open}
      onCancel={onCancel}
      width={520}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          Close
        </Button>,
        <Button key="send" type="primary" onClick={handleSend}>
          Send
        </Button>,
      ]}
    >
      <Form layout="vertical">
        {fields.map((field, index) => (
          <div key={field.key}>
            <Row gutter={8} align="bottom">
              <Col span={7}>
                <Form.Item label="Name" style={{ marginBottom: 0 }}>
                  <Input
                    value={field.name}
                    onChange={e =>
                      updateField(field.key, 'name', e.target.value)
                    }
                  />
                </Form.Item>
              </Col>
              <Col span={7}>
                <Form.Item
                  label="Target"
                  required
                  validateStatus={
                    errorKeys.includes(field.key) ? 'error' : undefined
                  }
                  help={errorKeys.includes(field.key) ? 'Required' : undefined}
                  style={{ marginBottom: 0 }}
                >
                  <Input
                    value={field.targetPort}
                    onChange={e =>
                      updateField(field.key, 'targetPort', e.target.value)
                    }
                  />
                </Form.Item>
              </Col>
              <Col span={7}>
                <Form.Item label="Desired" style={{ marginBottom: 0 }}>
                  <Input
                    value={field.desiredPort}
                    onChange={e =>
                      updateField(field.key, 'desiredPort', e.target.value)
                    }
                  />
                </Form.Item>
              </Col>
              <Col span={3} style={{ textAlign: 'center' }}>
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => removeField(field.key)}
                  disabled={fields.length === 1}
                />
              </Col>
            </Row>
            {index < fields.length - 1 && (
              <Divider style={{ margin: '16px 0' }} />
            )}
          </div>
        ))}
        <Form.Item style={{ textAlign: 'center', marginTop: 24 }}>
          <Button type="dashed" onClick={addField} disabled={isAddDisabled}>
            + Add Port
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  );
};
