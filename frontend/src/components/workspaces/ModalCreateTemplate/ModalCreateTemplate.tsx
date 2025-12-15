import type { FC } from 'react';
import { useState, useContext, useEffect } from 'react';
import { Modal, Form, Input, InputNumber, Select, Tooltip, Checkbox } from 'antd';
import { Button } from 'antd';
import type { CreateTemplateMutation } from '../../../generated-types';
import { InfoCircleOutlined } from '@ant-design/icons';
import type { RuleObject } from 'antd/es/form';

import {
  EnvironmentType,
  useWorkspaceTemplatesQuery,
  useImagesQuery,
  useWorkspaceSharedVolumesQuery,
} from '../../../generated-types';
import type { ApolloError, FetchResult } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { makeGuiSharedVolume } from '../../../utilsLogic';
import type { SharedVolume } from '../../../utils';
import { EnvironmentList } from './EnvironmentList';
import type { Image, Interval, TemplateForm } from './types';
import {
  formItemLayout,
  getDefaultTemplate,
  getImageLists,
  getImageNameNoVer,
  getImagesFromList,
  internalRegistry,
} from './utils';


export interface IModalCreateTemplateProps {
  workspaceNamespace: string;
  template?: TemplateForm;
  cpuInterval: Interval;
  ramInterval: Interval;
  diskInterval: Interval;
  show: boolean;
  setShow: (status: boolean) => void;
  submitHandler: (
    t: TemplateForm,
  ) => Promise<
    FetchResult<
      CreateTemplateMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  loading: boolean;
  isPersonal?: boolean;
}

const ModalCreateTemplate: FC<IModalCreateTemplateProps> = ({ ...props }) => {
  const {
    show,
    setShow,
    cpuInterval,
    ramInterval,
    diskInterval,
    template,
    submitHandler,
    loading,
    workspaceNamespace,
    isPersonal,
  } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);

  // Fetch all image lists
  const { data: dataImages } = useImagesQuery({
    variables: {},
    onError: apolloErrorCatcher,
  });

  const [form] = Form.useForm<TemplateForm>();

  // sharedVolumes must be declared at top-level (hooks cannot be conditional).
  const [sharedVolumes, setDataShVols] = useState<SharedVolume[]>([]);
  // Only fetch shared volumes when we have a valid namespace and the workspace is NOT personal.
  // Also limit fetching to when the modal is visible to avoid background/early fetches.
  const shouldFetchSharedVolumes =
    !!workspaceNamespace && isPersonal === false && !!show;

  useWorkspaceSharedVolumesQuery({
    variables: { workspaceNamespace },
    onError: apolloErrorCatcher,
    onCompleted: data =>
      setDataShVols(
        data.sharedvolumeList?.sharedvolumes
          ?.map(sv => makeGuiSharedVolume(sv))
          .sort((a, b) =>
            (a.prettyName ?? '').localeCompare(b.prettyName ?? ''),
          ) ?? [],
      ),
    fetchPolicy: 'network-only',
    skip: !shouldFetchSharedVolumes,
  });

  const validateName = async (_: unknown, name: string) => {
    if (!dataFetchTemplates || loadingFetchTemplates || errorFetchTemplates) {
      throw new Error('Error fetching templates');
    }

    if (!dataFetchTemplates.templateList) return;

    const trimmedName = name.trim().toLowerCase();
    const duplicateIndex = dataFetchTemplates.templateList.templates.findIndex(
      t => {
        return t?.spec?.prettyName?.toLowerCase() === trimmedName;
      },
    );

    if (duplicateIndex !== -1) {
      throw new Error(`This name has already been used in this workspace`);
    }
  };

  const fullLayout = {
    wrapperCol: { offset: 0, span: 24 },
  };

  const closehandler = () => {
    setShow(false);
  };

  const {
    data: dataFetchTemplates,
    error: errorFetchTemplates,
    loading: loadingFetchTemplates,
  } = useWorkspaceTemplatesQuery({
    onError: error => {
      console.error(
        'ModalCreateTemplate useWorkspaceTemplatesQuery error:',
        error,
        'workspaceNamespace:',
        workspaceNamespace,
      );
      apolloErrorCatcher(error);
    },
    variables: { workspaceNamespace },
  });

  const [availableImages, setAvailableImages] = useState<Image[]>([]);

  useEffect(() => {
    if (!dataImages) {
      setAvailableImages([]);
      return;
    }

    const imageLists = getImageLists(dataImages);
    const internalImages = imageLists.find(
      list => list.registryName === internalRegistry,
    );

    if (!internalImages) {
      setAvailableImages([]);
      return;
    }

    setAvailableImages(getImagesFromList(internalImages));
  }, [dataImages]);

  // Determine the final image URL
  const parseImage = (envType: EnvironmentType, image: string): string => {
    if (envType === EnvironmentType.VirtualMachine) {
      // For VMs, use the selected image from internal registry
      const selectedImage = availableImages.find(
        i => getImageNameNoVer(i.name) === image,
      );

      if (selectedImage) {
        return `${internalRegistry}/${selectedImage.name}`;
      }
    }

    // For other types, use the external image
    let finalImage = image;
    // If it doesn't include a registry, default to internal registry
    if (finalImage && !finalImage.includes('/') && !finalImage.includes('.')) {
      finalImage = `${internalRegistry}/${finalImage}`;
    }

    return finalImage;
  };

  const handleFormFinish = async (template: TemplateForm) => {
    const parsedTemplate = {
      ...template,
      inactivityTimeout: timeouts.inactivityTimeout.value === 0 ? 'never' : `${timeouts.inactivityTimeout.value}${timeouts.inactivityTimeout.unit}`,
      deleteAfter: timeouts.deleteAfter.value === 0 ? 'never' : `${timeouts.deleteAfter.value}${timeouts.deleteAfter.unit}`,
      environments: template.environments.map(env => ({
        ...env,
        image: parseImage(env.environmentType, env.image),
      })),
    };
    try {
      await submitHandler(parsedTemplate);
      
      setShow(false);
      form.resetFields();
      setTimeouts({
        inactivityTimeout: { value: 0, unit: '' },
        deleteAfter: { value: 0, unit: '' },
      });
    } catch (error) {
      console.error('ModalCreateTemplate submitHandler error:', error);
      apolloErrorCatcher(error as ApolloError);
    }
  };

  const getInitialValues = (template?: TemplateForm) => {
    if (template) return template;

    return getDefaultTemplate({
      cpu: cpuInterval,
      ram: ramInterval,
      disk: diskInterval,
    });
  };

  const handleFormSubmit = async () => {
    try {
      await form.validateFields();
    } catch (error) {
      console.error('ModalCreateTemplate validation error:', error);
    }
  };

  const TimeUnitOptions = [
    { label: 'Minutes', value: 'm' },
    { label: 'Hours', value: 'h' },
    { label: 'Days', value: 'd' },
  ];

  const [timeouts, setTimeouts] = useState({
    inactivityTimeout: { value: form.getFieldValue('inactivityTimeout') || 0, unit: '' },
    deleteAfter: { value: form.getFieldValue('deleteAfter') || 0, unit: '' },
  });

  const [automaticStoppingEnabled, setAutomaticStoppingEnabled] = useState(false);

  const handleTimeoutValueChange = (value: number | null, field: 'inactivityTimeout' | 'deleteAfter') => {
    setTimeouts(prevTimeouts => ({
      ...prevTimeouts,
      [field]: {
        value: value ? Number(value) : 0,
        unit: prevTimeouts[field].unit,
      },
    }));
    form.setFieldValue(field, {
    value,
    unit: timeouts[field].unit,
    });
    form.validateFields(['inactivityTimeout', 'deleteAfter']).catch(() => {});
  }

  const handleTimeUnitChange = (value: string, field: 'inactivityTimeout' | 'deleteAfter') => {
    setTimeouts(prevTimeouts => ({
      ...prevTimeouts,
      [field]: {
        value: prevTimeouts[field].value,
        unit: value,
      },
    }));
    form.setFieldValue(field, {
    value: timeouts[field].value,
    unit: value,
    });
    form.validateFields(['inactivityTimeout', 'deleteAfter']).catch(() => {});
  }

  const isTimeUnitDisabled = (field: 'inactivityTimeout' | 'deleteAfter') => {
    return timeouts[field].value === 0;
  };
  
  const validateTimeout = async (_: RuleObject, _val: { value: number; unit: string } ) => {
    if (_val.value === undefined || _val.value === 0) {
      return true; 
    }

    if (TimeUnitOptions.map(option => option.value).includes(_val.unit) === false) {
      throw new Error("Insert a valid time unit");
    } 
    return true;
  };

  const validateTimeoutOrder = async (rule: RuleObject , _val: { value: number; unit: string } | undefined, field: 'inactivityTimeout' | 'deleteAfter') => {
    console.log("RULE:", rule);
    const toMinutes = (t: { value: number; unit: string } | undefined) => {
      if (!t) return undefined;
      if (t.value === 0) return Infinity;
      const u = String(t.unit || '').toLowerCase();
      const mul = u === 'h' ? 60 : u === 'd' ? 1440 : 1;
      return Number(t.value) * mul;
    };
    
    const current =  form.getFieldValue(field);
    const inactivity = field === 'inactivityTimeout' ? current : form.getFieldValue('inactivityTimeout') as { value: number; unit: string } | undefined;
    const deleteAfter = field === 'deleteAfter' ? current : form.getFieldValue('deleteAfter') as { value: number; unit: string } | undefined;

    if (!inactivity || !deleteAfter) return;

    const inactivityMin = toMinutes(inactivity);
    const deleteAfterMin = toMinutes(deleteAfter);

    if (deleteAfterMin === Infinity) return;

    if (typeof inactivityMin !== 'number' || typeof deleteAfterMin !== 'number') return;

    if (inactivityMin >= deleteAfterMin) {
      throw new Error('Inactivity must be smaller than Expiration');
    }
    return;
  };

  return (
    <Modal
      destroyOnHidden={true}
      styles={{ body: { paddingBottom: '5px' } }}
      centered
      footer={null}
      title={template ? 'Modify template' : 'Create a new template'}
      open={show}
      onCancel={closehandler}
      width="600px">
      <Form
        form={form}
        onFinish={handleFormFinish}
        onSubmitCapture={handleFormSubmit}
        initialValues={getInitialValues(template)}
      >
        <Form.Item
          {...fullLayout}
          name="name"
          className="mt-1"
          required
          validateTrigger="onChange"
          rules={[
            {
              required: true,
              message: 'Please enter template name',
            },
            {
              validator: validateName,
            },
          ]}
        >
          <Input placeholder="Insert template name" allowClear />
        </Form.Item>
        <Checkbox className="mb-4" checked={automaticStoppingEnabled} onChange={e => setAutomaticStoppingEnabled(e.target.checked)}>Enable automatic clean-up</Checkbox>
        
      <Form.Item
        hidden={!automaticStoppingEnabled}
        label="Max Inactivity"
        name="inactivityTimeout"
        required={isTimeUnitDisabled('inactivityTimeout') ? false : true}
        validateTrigger="onChange"
        rules={[{ validator: validateTimeout }, { validator: (rule, value) => validateTimeoutOrder(rule, value, 'inactivityTimeout') }]}
        {...formItemLayout}> 
        
        <div className="flex gap-4 items-center">
          <Tooltip title={<><p>Instances based on this template are stopped / deleted (based on their persistency) if they're not accessed within this time (in certain special cases, activity might not be correctly detected, see <a href='https://github.com/netgroup-polito/CrownLabs/blob/master/operators/pkg/instautoctrl/README.md#instance-inactive-termination-controller'>here</a> for further technical information).</p> <p> <b>Set 0 to disable the feature.</b></p></>}>
            <InfoCircleOutlined />
          </Tooltip>
          <InputNumber
            onChange={value => handleTimeoutValueChange(value, 'inactivityTimeout')}
            min={0}
            max={60}
            defaultValue={timeouts.inactivityTimeout.value}
          >
          </InputNumber>

          <Select
            onChange={value => handleTimeUnitChange(value, 'inactivityTimeout')}
            disabled={isTimeUnitDisabled('inactivityTimeout')}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </Form.Item>

      <Form.Item
        hidden={!automaticStoppingEnabled}
        label="Max Lifetime"
        name="deleteAfter"
        required={isTimeUnitDisabled('deleteAfter') ? false : true}
        validateTrigger="onChange"
        rules={[{ validator: validateTimeout }]}
        {...formItemLayout}> 
        
        <div className="flex gap-4 items-center">
          <Tooltip title={<><p>Time, since the creation, after which instances based on this template are automatically deleted. Users will be preemptively alerted through email to take actions.</p> <p><b>Set 0 to disable the feature.</b></p></>}>
          
            <InfoCircleOutlined />
          </Tooltip>
          <InputNumber
            onChange={value => handleTimeoutValueChange(value, 'deleteAfter')}
            min={0}
            max={60}
            defaultValue={timeouts.deleteAfter.value}
          >
          </InputNumber>

          <Select
            onChange={value => handleTimeUnitChange(value, 'deleteAfter')}
            disabled={isTimeUnitDisabled('deleteAfter')}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </Form.Item>

        
        
        <EnvironmentList
          availableImages={availableImages}
          resources={{
            cpu: cpuInterval,
            ram: ramInterval,
            disk: diskInterval,
          }}
          sharedVolumes={sharedVolumes}
          isPersonal={isPersonal === undefined ? false : isPersonal}
        />

        <div className="flex justify-end gap-2">
          <Button htmlType="submit" onClick={() => closehandler()}>
            Cancel
          </Button>

          <Form.Item shouldUpdate>
            {() => {
              const fieldsError = form.getFieldsError();
              const hasErrors = fieldsError.some(
                ({ errors }) => errors.length > 0,
              );

              return (
                <Button htmlType="submit" type="primary" disabled={hasErrors}>
                  {!loading && (template ? 'Modify' : 'Create')}
                </Button>
              );
            }}
          </Form.Item>
        </div>
      </Form>
    </Modal>
  );
};

export type { TemplateForm as Template };
export default ModalCreateTemplate;
