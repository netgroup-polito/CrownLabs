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
} from 'antd';
import { DeleteOutlined, LoadingOutlined } from '@ant-design/icons';
import type { RuleObject } from 'antd/lib/form';
import { useApplyInstanceMutation } from '../../../../generated-types';
import {
  buildPublicExposurePatch,
  type PublicExposure,
} from '../../../../utils';
import { Phase } from '../../../../generated-types';

interface IPublicExposureModalProps {
  open: boolean;
  onCancel: () => void;
  allowPublicExposure: boolean;
  existingExposure?: PublicExposure;
  instanceId: string;
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
  tenantNamespace,
  manager,
}) => {
  const [form] = Form.useForm<FormValues>();
  const [isUpdating, setIsUpdating] = useState(false);
  const [lastSentData, setLastSentData] = useState<string | null>(null);

  const [applyInstanceMutation, { loading, error }] =
    useApplyInstanceMutation();

  const getInitialPorts = useMemo((): PortField[] => {
    if (
      existingExposure?.ports &&
      Array.isArray(existingExposure.ports) &&
      existingExposure.ports.length > 0
    ) {
      return existingExposure.ports.map(p => ({
        name: p?.name || '',
        targetPort: p?.targetPort ? String(p.targetPort) : '',
        desiredPort: allowPublicExposure
          ? p?.port === '0' || !p?.port
            ? ''
            : String(p.port)
          : '',
        protocol: ['TCP', 'UDP', 'SCTP'].includes(
          (p as { protocol?: string })?.protocol ?? '',
        )
          ? ((p as { protocol?: string })?.protocol ?? 'TCP')
          : 'TCP',
        _displayActualPort: p?.port && p.port !== '0' ? String(p.port) : '',
      }));
    }
    return [
      {
        name: '',
        targetPort: '',
        desiredPort: '',
        protocol: 'TCP',
        _displayActualPort: '',
      },
    ];
  }, [existingExposure, allowPublicExposure]);

  useEffect(() => {
    if (open) {
      const initialPorts = getInitialPorts;
      form.setFieldsValue({ ports: initialPorts });
    }
  }, [open, form, getInitialPorts]);

  useEffect(() => {
    if (open && existingExposure && !isUpdating) {
      const currentPorts = form.getFieldValue('ports') || [];
      const newInitialPorts = getInitialPorts;

      const hasSignificantChanges =
        newInitialPorts.length !== currentPorts.length ||
        newInitialPorts.some((np, index) => {
          const cp = currentPorts[index];
          return (
            !cp ||
            cp.targetPort !== np.targetPort ||
            cp.name !== np.name ||
            cp.protocol !== np.protocol ||
            cp.desiredPort !== np.desiredPort ||
            cp._displayActualPort !== np._displayActualPort
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
      
    return targetPorts.filter((port, index) => 
      targetPorts.indexOf(port) !== index
    );
  }, [ports]);

  const addButtonText =
    !ports || ports.length === 0 ? 'Add Port' : '+ Add Port';

  const shouldShowDisableButton = () => {
    const hasExistingExposure =
      existingExposure &&
      ((existingExposure.externalIP && existingExposure.phase !== Phase.Off) ||
        (existingExposure.ports && existingExposure.ports.length > 0));
    const hasCurrentPorts =
      ports &&
      ports.length > 0 &&
      ports.some(
        p =>
          p?.targetPort &&
          p.targetPort.trim() !== '' &&
          !isNaN(parseInt(p.targetPort, 10)) &&
          parseInt(p.targetPort, 10) > 0,
      );

    return !hasCurrentPorts && hasExistingExposure;
  };

  const isSendDisabled =
    isUpdating ||
    duplicateTargetPorts.length > 0 ||
    (!hasValidPorts &&
      ports &&
      ports.length > 0 &&
      ports.some(p => p?.targetPort && p.targetPort.trim() !== ''));

  const getButtonText = () => {
    if (shouldShowDisableButton()) {
      return 'Disable Public Exposure';
    }
    return 'Send';
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
    const duplicateTargetPorts = targetPorts.filter((port, index) => 
      targetPorts.indexOf(port) !== index
    );

    if (duplicateTargetPorts.length > 0) {
      form.setFields([
        {
          name: ['ports'],
          errors: [`Cannot expose the same target port multiple times. Duplicate target ports: ${[...new Set(duplicateTargetPorts)].join(', ')}`],
        },
      ]);
      return;
    }

    const targetPortCounts = new Map<number, number>();
    validPorts.forEach(p => {
      const targetPort = parseInt(p?.targetPort || '0', 10);
      targetPortCounts.set(targetPort, (targetPortCounts.get(targetPort) || 0) + 1);
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
                const currentIndex = (targetPortIndexes.get(targetPort) || 0) + 1;
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
          p?.targetPort === value && index !== fieldKey
      ).length;

      if (duplicateCount > 0) {
        return Promise.reject(new Error('This target port is already in use'));
      }

      return Promise.resolve();
    },
    [form],
  );

  return (
    <Modal
      open={open}
      onCancel={onCancel}
      width={650}
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
      {error && (
        <Alert
          type="error"
          message={error.message}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {duplicateTargetPorts.length > 0 && (
        <Alert
          type="error"
          message={`Cannot expose the same target port multiple times. Duplicate target ports: ${[...new Set(duplicateTargetPorts)].join(', ')}`}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {!hasValidPorts && ports && ports.length > 0 && (
        <Alert
          type="warning"
          message="At least one port with a valid Target Port is required to enable public exposure, or remove all ports to disable it."
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
                      <Col span={allowPublicExposure ? 4 : 5}>
                        <Form.Item
                          {...restField}
                          name={[name, 'targetPort']}
                          label={index === 0 ? 'Target Port' : ''}
                          rules={[
                            { required: true, message: 'Required' },
                            { 
                              validator: (rule, value) => 
                                targetPortValidator(rule, value, name)
                            },
                          ]}
                          validateTrigger={['onChange', 'onBlur']}
                          hasFeedback={false}
                          help=""
                        >
                          <Input placeholder="8080" disabled={isUpdating} />
                        </Form.Item>
                      </Col>
                      <Col span={allowPublicExposure ? 3 : 4}>
                        <Form.Item
                          {...restField}
                          name={[name, 'protocol']}
                          label={index === 0 ? 'Protocol' : ''}
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
                        <Col span={3}>
                          <Form.Item
                            {...restField}
                            name={[name, 'desiredPort']}
                            label={index === 0 ? 'Desired Port' : ''}
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
                        <Col span={3}>
                          <Form.Item
                            {...restField}
                            name={[name, '_displayActualPort']}
                            label={index === 0 ? 'Actual Port' : ''}
                          >
                            <Input
                              placeholder="—"
                              disabled={true}
                              style={{
                                backgroundColor: '#f5f5f5',
                                color: '#8c8c8c',
                                cursor: 'not-allowed',
                              }}
                            />
                          </Form.Item>
                        </Col>
                      )}
                      <Col span={1} style={{ textAlign: 'center' }}>
                        <Form.Item label={index === 0 ? '\u00A0' : ''}>
                          <Button
                            type="text"
                            danger
                            icon={<DeleteOutlined />}
                            onClick={() => remove(name)}
                            disabled={isUpdating}
                          />
                        </Form.Item>
                      </Col>
                    </Row>
                    {index < fields.length - 1 && (
                      <Divider style={{ margin: '12px 0' }} />
                    )}
                  </div>
                ))}

                {(!ports || ports.length === 0) && (
                  <Alert
                    type="info"
                    message="No ports configured"
                    description="Add a port to expose your instance services to the external network. Specify the target port of your service and optionally a desired public port."
                    showIcon
                    style={{ marginBottom: 16, textAlign: 'left' }}
                  />
                )}

                <Form.Item style={{ textAlign: 'center', marginTop: 24 }}>
                  <Button
                    type="dashed"
                    onClick={() =>
                      add({
                        name: '',
                        targetPort: '',
                        desiredPort: '',
                        _displayActualPort: '',
                        protocol: 'TCP',
                      })
                    }
                    disabled={isAddDisabled || isUpdating}
                  >
                    {addButtonText}
                  </Button>
                </Form.Item>

                <div
                  style={{
                    marginTop: 16,
                    textAlign: 'left',
                    fontSize: '12px',
                    color: '#666',
                    lineHeight: '1.4',
                  }}
                >
                  {existingExposure?.externalIP &&
                  existingExposure.phase !== Phase.Off ? (
                    <div>External IP: {existingExposure.externalIP}</div>
                  ) : (
                    <div>No external IP assigned yet</div>
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
