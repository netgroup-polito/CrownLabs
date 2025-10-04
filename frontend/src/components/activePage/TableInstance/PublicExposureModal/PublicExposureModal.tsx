import { useState, useEffect, useMemo, useCallback } from 'react';
import type { FC } from 'react';
import {
  Modal,
  Form,
  Button,
  Alert,
  Spin,
  Row,
  Col,
  Input,
  Select,
  Divider,
  Typography,
  Tooltip,
} from 'antd';
import { DeleteOutlined, LoadingOutlined, InfoCircleOutlined } from '@ant-design/icons';
import type { RuleObject } from 'antd/lib/form';
import { useApplyInstanceMutation } from '../../../../generated-types';
import {
  buildPublicExposurePatch,
  type PublicExposure,
  type PortListItem,
} from '../../../../utils';
import { Phase } from '../../../../generated-types';
const { Text } = Typography;

interface IPublicExposureModalProps {
  open: boolean;
  onCancel: () => void;
  allowPublicExposure: boolean;
  existingExposure?: PublicExposure;
  instanceId: string;
  instancePrettyName: string;
  tenantNamespace: string;
  manager: string;
}

interface PortField {
  name?: string;
  targetPort: string;
  desiredPort?: string;
  protocol?: string;
  _displayActualPort?: string;
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
  instancePrettyName,
  tenantNamespace,
  manager,
}) => {
  const instanceName = instancePrettyName;
  const [form] = Form.useForm<FormValues>();
  const [isUpdating, setIsUpdating] = useState(false);
  const [lastSentData, setLastSentData] = useState<string | null>(null);

  const [applyInstanceMutation, { loading, error }] =
    useApplyInstanceMutation();

  const getInitialPorts = useCallback(
    (exposurePorts: PortListItem[]) => {
      return exposurePorts.map(p => {
        let desiredPort = '';

        // If this port was originally requested as Auto (spec.port = 0),
        // show "Auto" placeholder regardless of assigned port
        if (p.isAutoPort || p.specPort === 0) {
          desiredPort = '';
        } else {
          // If this port was specifically requested (spec.port ≠ 0),
          // show the originally requested port number
          desiredPort = String(p.specPort);
        }

        return {
          targetPort: String(p.targetPort), // Convert to string to match PortField.targetPort
          protocol: p.protocol as 'TCP' | 'UDP' | 'SCTP',
          name: p.name,
          desiredPort: allowPublicExposure ? desiredPort : p.port || '',
          _displayActualPort: p.port && p.port !== '0' ? String(p.port) : '',
        };
      });
    },
    [allowPublicExposure],
  );

  useEffect(() => {
    if (open) {
      const existingPorts = existingExposure?.ports || [];
      let initialPorts;

      if (existingPorts.length > 0) {
        // If there are existing ports, use them
        initialPorts = getInitialPorts(existingPorts);
      } else {
        // If no existing ports, start with one empty row
        initialPorts = [
          {
            name: '',
            targetPort: '',
            desiredPort: '',
            protocol: 'TCP',
            _displayActualPort: '',
          },
        ];
      }

      form.setFieldsValue({ ports: initialPorts });
    }
  }, [open, form, getInitialPorts, existingExposure?.ports]);

  useEffect(() => {
    if (open && existingExposure && !isUpdating) {
      const currentPorts = form.getFieldValue('ports') || [];
      const newInitialPorts = getInitialPorts(existingExposure.ports || []);

      const hasSignificantChanges =
        newInitialPorts.length !== currentPorts.length ||
        newInitialPorts.some((np, index) => {
          const cp = currentPorts[index];
          return (
            !cp ||
            cp.targetPort !== np.targetPort ||
            cp.name !== np.name ||
            cp.protocol !== np.protocol ||
            cp.desiredPort !== np.desiredPort
          );
        });

      if (hasSignificantChanges) {
        form.setFieldsValue({ ports: newInitialPorts });
      }
    }
  }, [
    existingExposure?.externalIP,
    existingExposure?.phase,
    existingExposure?.ports,
    existingExposure,
    open,
    form,
    isUpdating,
    getInitialPorts,
  ]);

  useEffect(() => {
    if (lastSentData && existingExposure && isUpdating) {
      const currentData = JSON.stringify(existingExposure);
      if (currentData !== lastSentData) {
        setIsUpdating(false);
        setLastSentData(null);
      }
    }
  }, [existingExposure, lastSentData, isUpdating]);

  const ports = Form.useWatch('ports', form);
  const lastTargetPort =
    ports && Array.isArray(ports) && ports.length > 0
      ? ports[ports.length - 1]?.targetPort
      : undefined;
  const isAddDisabled =
    ports &&
    Array.isArray(ports) &&
    ports.length > 0 &&
    (!lastTargetPort ||
      !/^\d+$/.test(lastTargetPort) ||
      parseInt(lastTargetPort, 10) === 0);

  const hasValidPorts =
    ports &&
    Array.isArray(ports) &&
    ports.some(
      p =>
        p?.targetPort &&
        p.targetPort.trim() !== '' &&
        !isNaN(parseInt(p.targetPort, 10)) &&
        parseInt(p.targetPort, 10) > 0,
    );

  const duplicateTargetPorts = useMemo(() => {
    if (!ports || !Array.isArray(ports)) return [];

    const targetPorts = ports
      .filter(p => p?.targetPort && p.targetPort.trim() !== '')
      .map(p => parseInt(p?.targetPort || '0', 10))
      .filter(port => !isNaN(port) && port > 0);

    return targetPorts.filter(
      (port, index) => targetPorts.indexOf(port) !== index,
    );
  }, [ports]);

  const duplicateRequestedPorts = useMemo(() => {
    if (!ports || !Array.isArray(ports) || !allowPublicExposure) return [];

    const requestedPorts = ports
      .filter(
        p =>
          p?.desiredPort &&
          p.desiredPort.trim() !== '' &&
          p.desiredPort.trim() !== '0',
      )
      .map(p => parseInt(p?.desiredPort || '0', 10))
      .filter(port => !isNaN(port) && port > 0);

    return requestedPorts.filter(
      (port, index) => requestedPorts.indexOf(port) !== index,
    );
  }, [ports, allowPublicExposure]);

  const addButtonText =
    !ports || ports.length === 0 ? 'Add Port' : '+ Add Port';

  const hasUnsavedChanges = useMemo(() => {
    if (!ports) return false;

    const currentPorts = ports || [];
    const initialPorts = getInitialPorts(existingExposure?.ports || []);

    if (currentPorts.length !== initialPorts.length) {
      return true;
    }

    return currentPorts.some((currentPort, index) => {
      const initialPort = initialPorts[index];
      if (!initialPort) return true;

      return (
        currentPort?.name !== initialPort?.name ||
        currentPort?.targetPort !== initialPort?.targetPort ||
        currentPort?.protocol !== initialPort?.protocol ||
        currentPort?.desiredPort !== initialPort?.desiredPort
      );
    });
  }, [ports, getInitialPorts, existingExposure?.ports]);

  const isSendDisabled =
    isUpdating ||
    duplicateTargetPorts.length > 0 ||
    duplicateRequestedPorts.length > 0 ||
    !hasUnsavedChanges ||
    (!hasValidPorts &&
      ports &&
      ports.length > 0 &&
      ports.some(p => p?.targetPort && p.targetPort.trim() !== ''));

  const getButtonText = () => {
    return 'Save';
  };

  const onFinish = async (values: FormValues) => {
    if (loading || isUpdating) {
      return;
    }

    const validPorts =
      values.ports?.filter(
        p =>
          p?.targetPort &&
          p.targetPort.trim() !== '' &&
          !isNaN(parseInt(p.targetPort, 10)) &&
          parseInt(p.targetPort, 10) > 0,
      ) || [];

    const targetPorts = validPorts.map(p => parseInt(p?.targetPort || '0', 10));
    const duplicateTargetPorts = targetPorts.filter(
      (port, index) => targetPorts.indexOf(port) !== index,
    );

    if (duplicateTargetPorts.length > 0) {
      form.setFields([
        {
          name: ['ports'],
          errors: [
            `Cannot expose the same internal port multiple times. Duplicate internal ports: ${[...new Set(duplicateTargetPorts)].join(', ')}`,
          ],
        },
      ]);
      return;
    }

    // Check for duplicate requested ports (only when allowPublicExposure is true)
    if (allowPublicExposure) {
      const requestedPorts = validPorts
        .filter(
          p =>
            p?.desiredPort &&
            p.desiredPort.trim() !== '' &&
            p.desiredPort.trim() !== '0',
        )
        .map(p => parseInt(p?.desiredPort || '0', 10))
        .filter(port => !isNaN(port) && port > 0);

      const duplicateRequestedPorts = requestedPorts.filter(
        (port, index) => requestedPorts.indexOf(port) !== index,
      );

      if (duplicateRequestedPorts.length > 0) {
        form.setFields([
          {
            name: ['ports'],
            errors: [
              `Cannot request the same port multiple times. Duplicate requested ports: ${[...new Set(duplicateRequestedPorts)].join(', ')}`,
            ],
          },
        ]);
        return;
      }
    }

    const targetPortCounts = new Map<number, number>();
    validPorts.forEach(p => {
      const targetPort = parseInt(p?.targetPort || '0', 10);
      targetPortCounts.set(
        targetPort,
        (targetPortCounts.get(targetPort) || 0) + 1,
      );
    });

    const targetPortIndexes = new Map<number, number>();

    const normalized =
      validPorts.length === 0
        ? []
        : validPorts.map(p => {
            const targetPort = parseInt(p?.targetPort || '0', 10);
            const protocol =
              p?.protocol && ['TCP', 'UDP', 'SCTP'].includes(p.protocol)
                ? p.protocol
                : 'TCP';

            let name: string;
            if (p?.name && p.name.trim() !== '') {
              name = p.name.trim();
            } else {
              const count = targetPortCounts.get(targetPort) || 1;
              if (count > 1) {
                const currentIndex =
                  (targetPortIndexes.get(targetPort) || 0) + 1;
                targetPortIndexes.set(targetPort, currentIndex);
                name = `port-${targetPort}-${currentIndex}`;
              } else {
                name = `port-${targetPort}`;
              }
            }

            if (allowPublicExposure) {
              const port =
                p?.desiredPort && p.desiredPort.trim() !== ''
                  ? parseInt(p.desiredPort, 10)
                  : 0;

              return { name, targetPort, port, protocol };
            }
            return { name, targetPort, port: 0, protocol };
          });

    try {
      const patchJson = buildPublicExposurePatch(normalized);
      const variables = { instanceId, tenantNamespace, patchJson, manager };

      setIsUpdating(true);
      setLastSentData(JSON.stringify(existingExposure));

      const result = await applyInstanceMutation({ variables });

      if (result.data) {
        if (normalized.length === 0) {
          setTimeout(() => {
            setIsUpdating(false);
            setLastSentData(null);
            onCancel();
          }, 1000);
        } else {
          setTimeout(() => {
            setIsUpdating(false);
            setLastSentData(null);
          }, 1000);
        }
      }
    } catch (_error) {
      setIsUpdating(false);
      setLastSentData(null);
    }
  };

  const portValidator = useCallback((_rule: RuleObject, value: string) => {
    if (!value) {
      return Promise.resolve();
    }
    const num = parseInt(value, 10);
    if (isNaN(num) || num <= 0 || num > 65535) {
      return Promise.reject(new Error('Port must be between 1 and 65535'));
    }
    return Promise.resolve();
  }, []);

  const targetPortValidator = useCallback(
    (_rule: RuleObject, value: string, fieldKey?: number) => {
      if (!value) {
        return Promise.resolve();
      }

      const num = parseInt(value, 10);
      if (isNaN(num) || num <= 0 || num > 65535) {
        return Promise.reject(new Error('Port must be between 1 and 65535'));
      }

      const currentPorts = form.getFieldValue('ports') || [];
      const duplicateCount = currentPorts.filter(
        (p: PortField, index: number) =>
          p?.targetPort === value && index !== fieldKey,
      ).length;

      if (duplicateCount > 0) {
        return Promise.reject(
          new Error('This internal port is already in use'),
        );
      }

      return Promise.resolve();
    },
    [form],
  );

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      width={680}
      title={
        <>
          Public Port Exposure for <em>{instanceName}</em>
        </>
      }
      footer={[
        <Button
          key="cancel"
          onClick={onCancel}
          disabled={loading || isUpdating}
        >
          Close
        </Button>,
        <Button
          key="send"
          type="primary"
          onClick={() => form.submit()}
          loading={loading || isUpdating}
          disabled={isSendDisabled}
        >
          {getButtonText()}
        </Button>,
      ]}
    >
      <Spin
        spinning={isUpdating}
        tip="Updating ports..."
        indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />}
      >
        <Form
          form={form}
          name={`dynamic_port_form_${instanceId}`}
          onFinish={onFinish}
          autoComplete="off"
          layout="vertical"
          scrollToFirstError={{ behavior: 'smooth', block: 'center' }}
        >
          <Form.List name="ports">
            {(fields, { add, remove }) => {
              // Function to add a port and scroll to it using Antd native API
              const handleAddPort = () => {
                const newIndex = fields.length;
                add({
                  name: '',
                  targetPort: '',
                  desiredPort: '',
                  _displayActualPort: '',
                  protocol: 'TCP',
                });
                
                // Use Antd's scrollToField to scroll to the newly added field
                setTimeout(() => {
                  form.scrollToField(['ports', newIndex, 'targetPort'], {
                    behavior: 'smooth',
                    block: 'center',
                  });
                }, 100); // Small delay to ensure the new element is rendered
              };

              return (
              <>
                <div
                  className="ant-modal-body"
                  style={{
                    maxHeight: 320,
                    overflowY: 'auto',
                    paddingRight: 8,
                    marginBottom: 16,
                  }}
                >
                  {fields.map(({ key, name, ...restField }, index) => (
                    <div key={key}>
                      <Row gutter={8} align="bottom">
                        <Col span={allowPublicExposure ? 4 : 5}>
                          <Form.Item
                            {...restField}
                            name={[name, 'name']}
                            label={index === 0 ? 'Name' : ''}
                            rules={[{ required: false }]}
                          >
                            <Input
                              placeholder="Port name"
                              disabled={isUpdating}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={allowPublicExposure ? 5 : 6}>
                          <Form.Item
                            {...restField}
                            name={[name, 'targetPort']}
                            label={
                              index === 0 ? (
                                <span>
                                  Internal Port{' '}
                                  <Tooltip title="The port number inside your container where your service is listening (e.g., 8080 for a web server)">
                                    <InfoCircleOutlined style={{ color: '#1890ff', fontSize: '10px', marginLeft: '2px' }} />
                                  </Tooltip>
                                </span>
                              ) : (
                                ''
                              )
                            }
                            rules={[
                              { required: true, message: 'Required' },
                              {
                                validator: (rule, value) =>
                                  targetPortValidator(rule, value, name),
                              },
                            ]}
                            validateTrigger={['onChange', 'onBlur']}
                            hasFeedback={false}
                            help=""
                          >
                            <Input
                              placeholder="e.g. 8080"
                              disabled={isUpdating}
                            />
                          </Form.Item>
                        </Col>
                        <Col span={allowPublicExposure ? 4 : 6}>
                          <Form.Item
                            {...restField}
                            name={[name, 'protocol']}
                            label={
                              index === 0 ? (
                                <span>
                                  Protocol{' '}
                                  <Tooltip title="The network protocol used for communication. TCP is most common for web services, UDP for real-time applications">
                                    <InfoCircleOutlined style={{ color: '#1890ff', fontSize: '10px', marginLeft: '2px' }} />
                                  </Tooltip>
                                </span>
                              ) : (
                                ''
                              )
                            }
                            rules={[{ required: true, message: 'Required' }]}
                            initialValue="TCP"
                          >
                            <Select
                              placeholder="Select protocol"
                              disabled={isUpdating}
                            >
                              <Select.Option value="TCP">TCP</Select.Option>
                              <Select.Option value="UDP">UDP</Select.Option>
                              <Select.Option value="SCTP">SCTP</Select.Option>
                            </Select>
                          </Form.Item>
                        </Col>
                        {allowPublicExposure && (
                          <Col span={5}>
                            <Form.Item
                              {...restField}
                              name={[name, 'desiredPort']}
                              label={
                                index === 0 ? (
                                  <span>
                                    Requested Port{' '}
                                    <Tooltip title="Specific external port you want to use. Leave empty for automatic assignment by the system">
                                      <InfoCircleOutlined style={{ color: '#1890ff', fontSize: '10px', marginLeft: '2px' }} />
                                    </Tooltip>
                                  </span>
                                ) : (
                                  ''
                                )
                              }
                              rules={[{ validator: portValidator }]}
                              validateTrigger={['onChange', 'onBlur']}
                              hasFeedback={false}
                              help=""
                            >
                              <Input placeholder="Auto" disabled={isUpdating} />
                            </Form.Item>
                          </Col>
                        )}
                        {allowPublicExposure && (
                          <Col span={5}>
                            <Form.Item
                              {...restField}
                              name={[name, '_displayActualPort']}
                              label={
                                index === 0 ? (
                                  <span>
                                    Assigned Port{' '}
                                    <Tooltip title="The actual external port assigned by the system. This is the port you'll use to access your service from outside">
                                      <InfoCircleOutlined style={{ color: '#1890ff', fontSize: '10px', marginLeft: '2px' }} />
                                    </Tooltip>
                                  </span>
                                ) : (
                                  ''
                                )
                              }
                            >
                              <Input placeholder="—" disabled={true} />
                            </Form.Item>
                          </Col>
                        )}
                        <Col span={1}>
                          <Form.Item label={index === 0 ? '\u00A0' : ''}>
                            <Button
                              type="text"
                              danger
                              icon={<DeleteOutlined />}
                              onClick={() => remove(name)}
                              disabled={isUpdating}
                              block
                            />
                          </Form.Item>
                        </Col>
                      </Row>
                      {index < fields.length - 1 && (
                        <Divider style={{ margin: '8px 0' }} />
                      )}
                    </div>
                  ))}
                </div>

                {(!ports || ports.length === 0) && (
                  <Alert
                    type="info"
                    message="No ports configured"
                    description="Add a port to expose your instance services to the external network. Specify the internal port of your service and optionally a request port. Note: Saving without any valid ports will remove the external IP address if one is currently assigned."
                    showIcon
                  />
                )}

                <Form.Item style={{ textAlign: 'center', marginTop: '24px' }}>
                  <Button
                    type="dashed"
                    onClick={handleAddPort}
                    disabled={isAddDisabled || isUpdating}
                  >
                    {addButtonText}
                  </Button>
                </Form.Item>

                {error && (
                  <Alert type="error" message={error.message} showIcon />
                )}

                {!hasValidPorts && ports && ports.length > 0 && (
                  <Alert
                    type="warning"
                    message="At least one port with a valid Internal value is required to enable public exposure, or remove all ports to disable it."
                    showIcon
                  />
                )}

                {duplicateTargetPorts.length > 0 && (
                  <Alert
                    type="error"
                    message={`Cannot expose the same internal port multiple times. Duplicate internal ports: ${[...new Set(duplicateTargetPorts)].join(', ')}`}
                    showIcon
                  />
                )}

                {duplicateRequestedPorts.length > 0 && (
                  <Alert
                    type="error"
                    message={`Cannot request the same port multiple times. Duplicate requested ports: ${[...new Set(duplicateRequestedPorts)].join(', ')}`}
                    showIcon
                  />
                )}

                <Text type="secondary" style={{ fontSize: 12 }}>
                  {existingExposure?.externalIP &&
                  existingExposure.phase !== Phase.Off ? (
                    <span
                      style={{
                        textDecoration: !hasValidPorts
                          ? 'line-through'
                          : 'none',
                        color: !hasValidPorts ? '#ff4d4f' : undefined,
                      }}
                    >
                      External IP: {existingExposure.externalIP}
                    </span>
                  ) : (
                    <span>No external IP assigned yet</span>
                  )}
                </Text>
              </>
              );
            }}
          </Form.List>
        </Form>
      </Spin>
    </Modal>
  );
};
