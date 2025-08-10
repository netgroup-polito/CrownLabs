import {
  Modal,
  Button,
  Form,
  Input,
  Row,
  Col,
  Divider,
  Alert,
  type FormRule,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';

import { type FC, useCallback, useEffect, useMemo } from 'react';
import type { PublicExposure } from '../../../../utils';
import { useApplyInstanceMutation } from '../../../../generated-types';
import { buildPublicExposurePatch } from '../../../../utils';

interface IPublicExposureModalProps {
  open: boolean;
  onCancel: () => void;
  allowPublicExposure: boolean;
  
  existingExposure?: PublicExposure;
  // k8s patch context
  instanceId: string;
  tenantNamespace: string;
  manager: string;
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
  allowPublicExposure,
  existingExposure,
  instanceId,
  tenantNamespace,
  manager,
}) => {
  const [form] = Form.useForm<FormValues>();
  // GraphQL mutation hook for applyInstance
  const [applyInstanceMutation, { loading, error }] =
    useApplyInstanceMutation();
  
  // Create initial ports from existing exposure
  const getInitialPorts = useMemo((): PortField[] => {
    if (existingExposure?.ports && existingExposure.ports.length > 0) {
      return existingExposure.ports.map(p => ({
        name: p.name || '',
        targetPort: String(p.targetPort),
        desiredPort: allowPublicExposure ? String(p.port) : '',
      }));
    }
    return [{ name: '', targetPort: '', desiredPort: '' }];
  }, [existingExposure, allowPublicExposure]);

  // Reset form values when modal opens or existingExposure changes
  useEffect(() => {
    if (open) {
      // Use setTimeout to ensure the form is properly initialized
      setTimeout(() => {
        form.setFieldsValue({ ports: getInitialPorts });
      }, 0);
    }
  }, [open, existingExposure, allowPublicExposure, form, getInitialPorts]);

  // Additional effect to update form when existingExposure changes while modal is open
  useEffect(() => {
    if (open && existingExposure) {
      // Only update if ports actually changed to avoid infinite loops
      const currentPorts = form.getFieldValue('ports') || [];
      const portsChanged = JSON.stringify(currentPorts) !== JSON.stringify(getInitialPorts);
      
      if (portsChanged) {
        form.setFieldsValue({ ports: getInitialPorts });
      }
    }
  }, [existingExposure, open, form, getInitialPorts]);

  const ports = Form.useWatch('ports', form);
  const lastTargetPort = ports?.[ports.length - 1]?.targetPort;
  const isAddDisabled =
    !lastTargetPort ||
    !/^\d+$/.test(lastTargetPort) ||
    parseInt(lastTargetPort, 10) === 0;

  const onFinish = async (values: FormValues) => {
    const normalized = values.ports.map(p => {
      const targetPort = parseInt(p.targetPort, 10);
      if (allowPublicExposure) {
        return { name: p.name, targetPort, port: p.desiredPort || '0' };
      }
      return { name: p.name, targetPort };
    });
    
    try {
      // build patch for publicExposure via helper
      const patchJson = buildPublicExposurePatch(normalized);
      const variables = { instanceId, tenantNamespace, patchJson, manager };
      
      const result = await applyInstanceMutation({ variables });
      
      // If mutation successful
      if (result.data) {
        onCancel();
      }
    } catch (error) {
      console.error('❌ Backend error:', error);
      // error displayed via Alert
    }
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
        <Button key="cancel" onClick={onCancel} disabled={loading}>
          Close
        </Button>,
        <Button
          key="send"
          type="primary"
          onClick={() => form.submit()}
          loading={loading}
        >
          Send
        </Button>,
      ]}
    >
      {error && (
        <Alert
          type="error"
          message={error.message}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      <Form
        form={form}
        name="dynamic_port_form"
        onFinish={onFinish}
        autoComplete="off"
        layout="vertical"
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
                    {allowPublicExposure && (
                      <Col span={7}>
                        <Form.Item
                          {...restField}
                          name={[name, 'desiredPort']}
                          label="Desired Port"
                          style={{ marginBottom: 8 }}
                          rules={[{ validator: portValidator }]}
                        >
                          <Input placeholder="auto-assign (leave empty)" />
                        </Form.Item>
                      </Col>
                    )}
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
