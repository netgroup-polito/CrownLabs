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
  // Campo virtuale solo per visualizzazione - non viene inviato al backend
  _displayActualPort?: string;
}

interface FormValues {
  ports: PortField[];
}

// Removed FormRule interface, use antd's RuleObject type for validator

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
    console.log('🔍 Debug - getInitialPorts called with existingExposure:', existingExposure);
    
    if (
      existingExposure?.ports &&
      Array.isArray(existingExposure.ports) &&
      existingExposure.ports.length > 0
    ) {
      const mappedPorts = existingExposure.ports.map(p => ({
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
        // Campo virtuale per mostrare la porta effettivamente assegnata dal backend status
        _displayActualPort: p?.port && p.port !== '0' ? String(p.port) : '',
      }));
      
      console.log('🔍 Debug - Mapped ports:', mappedPorts);
      return mappedPorts;
    }
    
    console.log('🔍 Debug - No existing ports, returning empty port');
    // Se non c'è esistingExposure o non ha porte, inizializza con una porta vuota
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

  // Inizializza il form solo all'apertura del modal e quando cambia existingExposure significativamente
  useEffect(() => {
    if (open) {
      const initialPorts = getInitialPorts;
      form.setFieldsValue({ ports: initialPorts });
      console.log('🔄 Form initialized with ports:', initialPorts);
    }
  }, [open, form, getInitialPorts]);

  // Aggiorna il form solo quando esistingExposure cambia e non stiamo aggiornando
  useEffect(() => {
    if (open && existingExposure && !isUpdating) {
      const currentPorts = form.getFieldValue('ports') || [];
      const newInitialPorts = getInitialPorts;

      console.log('🔍 Debug - Current ports in form:', currentPorts);
      console.log('🔍 Debug - New initial ports from existingExposure:', newInitialPorts);
      console.log('🔍 Debug - existingExposure:', existingExposure);

      // Confronta solo se c'è una vera differenza nei dati essenziali
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

      console.log('🔍 Debug - Has significant changes:', hasSignificantChanges);

      if (hasSignificantChanges) {
        console.log('📝 Updating form due to external changes');
        console.log('🔍 Debug - Overwriting form with:', newInitialPorts);
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

  // Gestisci il completamento delle operazioni
  useEffect(() => {
    if (lastSentData && existingExposure && isUpdating) {
      const currentData = JSON.stringify(existingExposure);
      if (currentData !== lastSentData) {
        console.log('✅ Update completed, resetting state');
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

  // Check if we have any valid ports for Send button
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

  // Aggiorna la logica per il pulsante "Add Port" - rimuove variabile non utilizzata
  const addButtonText =
    !ports || ports.length === 0 ? 'Add Port' : '+ Add Port';

  // Determina se dovremmo mostrare "Disable Public Exposure"
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

    // Mostra "Disable" solo se:
    // 1. Non ci sono porte valide nel form corrente E
    // 2. C'è un'esposizione esistente attiva (IP esterno o porte)
    return !hasCurrentPorts && hasExistingExposure;
  };

  // Determina se il pulsante Send dovrebbe essere disabilitato
  const isSendDisabled =
    isUpdating ||
    (!hasValidPorts &&
      ports &&
      ports.length > 0 &&
      ports.some(p => p?.targetPort && p.targetPort.trim() !== ''));

  // Determina il testo del pulsante
  const getButtonText = () => {
    if (shouldShowDisableButton()) {
      return 'Disable Public Exposure';
    }
    return 'Send';
  };

  const onFinish = async (values: FormValues) => {
    if (loading || isUpdating) {
      console.log('⚠️ Operation already in progress, ignoring...');
      return;
    }

    console.log('🔍 Debug - Form values received:', values);

    const validPorts =
      values.ports?.filter(
        p =>
          p?.targetPort &&
          p.targetPort.trim() !== '' &&
          !isNaN(parseInt(p.targetPort, 10)) &&
          parseInt(p.targetPort, 10) > 0,
      ) || [];

    console.log('🔍 Debug - Valid ports after filtering:', validPorts);

    // Create a map to track targetPort counts for unique naming
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

            // Generate unique name if not provided
            let name: string;
            if (p?.name && p.name.trim() !== '') {
              name = p.name.trim();
            } else {
              const count = targetPortCounts.get(targetPort) || 1;
              if (count > 1) {
                // Multiple ports with same targetPort, add index
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

    console.log('🔍 Debug - Normalized ports to send:', normalized);

    try {
      const patchJson = buildPublicExposurePatch(normalized);
      const variables = { instanceId, tenantNamespace, patchJson, manager };

      console.log('📦 Sending patch:', patchJson);
      console.log('🔍 Debug - Variables:', variables);

      // Imposta il flag di aggiornamento prima della mutazione
      setIsUpdating(true);
      setLastSentData(JSON.stringify(existingExposure));

      const result = await applyInstanceMutation({ variables });

      console.log('🔍 Debug - Mutation result:', result);

      if (result.data) {
        console.log('✅ Public exposure updated successfully');
        console.log('🔍 Debug - Result data:', result.data);

        // Non chiudere il modal automaticamente per permettere ulteriori modifiche
        if (normalized.length === 0) {
          // Per la disabilitazione, chiudi il modal dopo un breve delay
          setTimeout(() => {
            setIsUpdating(false);
            setLastSentData(null);
            onCancel();
          }, 1000);
        } else {
          // Per l'abilitazione, mantieni il modal aperto
          setTimeout(() => {
            setIsUpdating(false);
            setLastSentData(null);
          }, 1000);
        }
      }
    } catch (error) {
      console.error('❌ Backend error:', error);
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
                            { validator: portValidator },
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
