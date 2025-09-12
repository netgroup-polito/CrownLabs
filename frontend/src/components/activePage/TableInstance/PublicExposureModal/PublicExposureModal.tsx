import {
  Modal,
  Button,
  Form,
  Input,
  Row,
  Col,
  Divider,
  Alert,
  Radio,
  Spin,
  type FormRule,
} from 'antd';
import { DeleteOutlined, LoadingOutlined } from '@ant-design/icons';

import { type FC, useCallback, useEffect, useMemo, useState } from 'react';
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
  protocol?: string;
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
  // Internal loading state for real-time updates
  const [isUpdating, setIsUpdating] = useState(false);
  // Store the last sent data to detect when updates are complete
  const [lastSentData, setLastSentData] = useState<string | null>(null);
  
  // GraphQL mutation hook for applyInstance
  const [applyInstanceMutation, { loading, error }] =
    useApplyInstanceMutation();

  // Create initial ports from existing exposure
  const getInitialPorts = useMemo((): PortField[] => {
    // Safe check for existing exposure and ports array
    if (existingExposure?.ports && Array.isArray(existingExposure.ports) && existingExposure.ports.length > 0) {
      return existingExposure.ports.map(p => ({
        name: p?.name || '',
        targetPort: p?.targetPort ? String(p.targetPort) : '',
        // For Desired Port: show the original requested value
        // If the port is auto-assigned (was 0 or empty), leave empty to indicate auto-assignment
        desiredPort: allowPublicExposure 
          ? (p?.port === '0' || !p?.port ? '' : String(p.port))
          : '',
        // Protocol: use the value from backend, default to TCP if not set, support SCTP
        protocol: (p as any)?.protocol && ['TCP', 'UDP', 'SCTP'].includes((p as any).protocol) 
          ? (p as any).protocol 
          : 'TCP',
      }));
    }
    // If no existing exposure or no ports, start with a single empty port
    return [{ name: '', targetPort: '', desiredPort: '', protocol: 'TCP' }];
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
      try {
        // Only update if ports actually changed to avoid infinite loops
        const currentPorts = form.getFieldValue('ports') || [];
        const newInitialPorts = getInitialPorts;
        const portsChanged =
          JSON.stringify(currentPorts) !== JSON.stringify(newInitialPorts);

        if (portsChanged) {
          form.setFieldsValue({ ports: newInitialPorts });
        }
      } catch (error) {
        console.error('Error updating form values:', error);
        // Fallback to safe default
        form.setFieldsValue({ ports: [{ name: '', targetPort: '', desiredPort: '', protocol: 'TCP' }] });
      }
    }
  }, [existingExposure, open, form, getInitialPorts]);

  // Effect to manage loading state based on data updates
  useEffect(() => {
    if (lastSentData && existingExposure) {
      const currentData = JSON.stringify(existingExposure);
      if (currentData !== lastSentData) {
        // Data has been updated, stop loading
        setIsUpdating(false);
        setLastSentData(null);
      }
    }
  }, [existingExposure, lastSentData]);

  const ports = Form.useWatch('ports', form);
  const lastTargetPort = ports && Array.isArray(ports) && ports.length > 0 ? ports[ports.length - 1]?.targetPort : undefined;
  const isAddDisabled =
    ports && Array.isArray(ports) && ports.length > 0 && (
      !lastTargetPort ||
      !/^\d+$/.test(lastTargetPort) ||
      parseInt(lastTargetPort, 10) === 0
    );

  // Check if we have any valid ports for Send button
  const hasValidPorts = ports && Array.isArray(ports) && ports.some(p => 
    p?.targetPort && p.targetPort.trim() !== '' && 
    !isNaN(parseInt(p.targetPort, 10)) && 
    parseInt(p.targetPort, 10) > 0
  );

  const isSendDisabled = isUpdating || !hasValidPorts;

  const onFinish = async (values: FormValues) => {
    // Prevent multiple submissions
    if (loading) {
      console.log('⚠️ Submission already in progress, ignoring...');
      return;
    }

    // Backend cannot handle empty ports, so prevent submission
    if (!values.ports || !Array.isArray(values.ports) || values.ports.length === 0) {
      console.log('❌ No ports to submit - backend requires at least one port');
      return;
    }

    // Filter out empty ports (ports without targetPort)
    const validPorts = values.ports.filter(p => 
      p?.targetPort && p.targetPort.trim() !== '' && 
      !isNaN(parseInt(p.targetPort, 10)) && 
      parseInt(p.targetPort, 10) > 0
    );

    if (validPorts.length === 0) {
      console.log('❌ No valid ports to submit - backend requires at least one valid port');
      return;
    }

    const normalized = validPorts.map(p => {
      const targetPort = parseInt(p?.targetPort || '0', 10);
      // Ensure protocol is valid according to CRD (TCP, UDP, SCTP)
      const protocol = p?.protocol && ['TCP', 'UDP', 'SCTP'].includes(p.protocol) ? p.protocol : 'TCP';
      
      if (allowPublicExposure) {
        // Ensure name is provided (required by CRD)
        const name = p?.name && p.name.trim() !== '' ? p.name.trim() : `port-${targetPort}`;
        // Parse port or set to 0 for auto-assignment (required by CRD)
        const port = p?.desiredPort && p.desiredPort.trim() !== '' ? parseInt(p.desiredPort, 10) : 0;
        
        return { 
          name, 
          targetPort, 
          port,
          protocol
        };
      }
      // For non-public exposure, still need required fields
      const name = p?.name && p.name.trim() !== '' ? p.name.trim() : `port-${targetPort}`;
      return { 
        name, 
        targetPort, 
        port: 0,
        protocol
      };
    });

    try {
      // build patch for publicExposure via helper
      const patchJson = buildPublicExposurePatch(normalized);
      const variables = { instanceId, tenantNamespace, patchJson, manager };

      // Set loading state and store current data to detect updates
      setIsUpdating(true);
      setLastSentData(JSON.stringify(existingExposure));

      const result = await applyInstanceMutation({ variables });

      // If mutation successful, keep modal open to see real-time updates
      if (result.data) {
        console.log('✅ Public exposure updated successfully');
        // Loading state will be cleared by the effect when data updates
      }
    } catch (error) {
      console.error('❌ Backend error:', error);
      // Clear loading state on error
      setIsUpdating(false);
      setLastSentData(null);
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
      width={650}
      footer={[
        <Button key="cancel" onClick={onCancel} disabled={loading || isUpdating}>
          Close
        </Button>,
        <Button
          key="send"
          type="primary"
          onClick={() => form.submit()}
          loading={loading}
          disabled={isSendDisabled}
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
      
      {!hasValidPorts && (
        <Alert
          type="info"
          message="At least one port with a valid Target Port is required to enable public exposure."
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      
      <Spin 
        spinning={isUpdating} 
        tip="Updating ports..."
        indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />}
      >
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
                    <Col span={4}>
                      <Form.Item
                        {...restField}
                        name={[name, 'name']}
                        label="Name"
                        style={{ marginBottom: 8 }}
                      >
                        <Input placeholder="web" size="small" />
                      </Form.Item>
                    </Col>
                    <Col span={4}>
                      <Form.Item
                        {...restField}
                        name={[name, 'targetPort']}
                        label={<span style={{ whiteSpace: 'nowrap' }}>Target Port (internal)</span>}
                        style={{ marginBottom: 8 }}
                        rules={[
                          { required: true, message: 'Required' },
                          { validator: portValidator },
                        ]}
                      >
                        <Input placeholder="8080" size="small" />
                      </Form.Item>
                    </Col>
                    <Col span={3}>
                      <Form.Item
                        {...restField}
                        name={[name, 'protocol']}
                        label={<span style={{ whiteSpace: 'nowrap' }}>Protocol</span>}
                        style={{ marginBottom: 8 }}
                        initialValue="TCP"
                      >
                        <Radio.Group 
                          size="small" 
                          style={{ 
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '1px',
                            fontSize: '9px'
                          }}
                        >
                          <Radio value="TCP" style={{ fontSize: '9px', margin: 0, lineHeight: '12px' }}>TCP</Radio>
                          <Radio value="UDP" style={{ fontSize: '9px', margin: 0, lineHeight: '12px' }}>UDP</Radio>
                        </Radio.Group>
                      </Form.Item>
                    </Col>
                    {allowPublicExposure && (
                      <Col span={4}>
                        <Form.Item
                          {...restField}
                          name={[name, 'desiredPort']}
                          label={<span style={{ whiteSpace: 'nowrap' }}>Desired Port (public)</span>}
                          style={{ marginBottom: 8 }}
                          rules={[{ validator: portValidator }]}
                        >
                          <Input placeholder="auto" size="small" />
                        </Form.Item>
                      </Col>
                    )}
                    {allowPublicExposure && (
                      <Col span={4}>
                        <Form.Item
                          label={<span style={{ whiteSpace: 'nowrap' }}>Actual Port (public)</span>}
                          style={{ marginBottom: 8 }}
                        >
                          <Input 
                            value={(() => {
                              // Safe access to the actual assigned port from existing exposure
                              const ports = existingExposure?.ports;
                              if (!ports || !Array.isArray(ports) || index >= ports.length) {
                                return 'N/A';
                              }
                              
                              const actualPort = ports[index]?.port;
                              
                              // Show actual port if assigned and not 0, otherwise show "N/A"
                              if (actualPort && actualPort !== '0' && String(actualPort).trim() !== '') {
                                return String(actualPort);
                              }
                              
                              return 'N/A';
                            })()} 
                            disabled 
                            size="small"
                            style={{ 
                              backgroundColor: '#f5f5f5',
                              color: '#666',
                              cursor: 'not-allowed',
                              fontSize: '11px'
                            }}
                          />
                        </Form.Item>
                      </Col>
                    )}
                    <Col
                      span={1}
                      style={{ textAlign: 'center', paddingBottom: '8px' }}
                    >
                      <Button
                        type="text"
                        danger
                        icon={<DeleteOutlined />}
                        onClick={() => remove(name)}
                        size="small"
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
              
              {/* IP Address Display */}
              <div style={{ 
                marginTop: 16, 
                textAlign: 'left',
                fontSize: '12px',
                color: '#666',
                lineHeight: '1.4'
              }}>
                {existingExposure?.externalIP ? (
                  <div>
                    <strong>External IP:</strong> {existingExposure.externalIP}
                  </div>
                ) : (
                  <div>
                    <div style={{ color: '#999', marginBottom: '4px' }}>
                      External IP not yet synchronized in instance status
                    </div>
                    <div style={{ color: '#1890ff', fontSize: '11px' }}>
                      💡 Check LoadBalancer service: <code>{instanceId}-pe</code>
                    </div>
                  </div>
                )}
              </div>
            </>
          )}
        </Form.List>
      </Form>
      </Spin>
    </Modal>
  );
};
