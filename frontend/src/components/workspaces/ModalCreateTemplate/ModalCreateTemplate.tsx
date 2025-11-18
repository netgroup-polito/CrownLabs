import type { FC } from 'react';
import { useState, useContext, useEffect } from 'react';
import { Modal, Form, Input, Select, Tooltip } from 'antd';
import { Button } from 'antd';
import type { CreateTemplateMutation } from '../../../generated-types';
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
import { InfoCircleFilled  } from '@ant-design/icons';
import {
  formItemLayout,
  getDefaultTemplate,
  getImageLists,
  getImageNameNoVer,
  getImagesFromList,
  internalRegistry,
} from './utils';
import { validate } from 'graphql';

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
  inactivityTimeout: string;
  deleteAfter: string;
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
    inactivityTimeout,
    deleteAfter,
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
      environments: template.environments.map(env => ({
        ...env,
        image: parseImage(env.environmentType, env.image),
      })),
    };
    try {
      await submitHandler(parsedTemplate);
      console.log('Submitting template:', parsedTemplate);

      //setShow(false);
      form.resetFields();
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

// Time unit options
const timeUnitOptions = [
  { value: 'm', label: 'Minutes' },
  { value: 'h', label: 'Hours' },
  { value: 'd', label: 'Days' },
];

  const [timeouts, setTimeouts] = useState<{
  deleteAfterUnit: string;
  inactivityTimeoutUnit: string;
  deleteAfterValue: string;
  inactivityTimeoutValue: string;
}>({
  deleteAfterUnit:
    deleteAfter && String(deleteAfter) !== 'Never'
      ? String(deleteAfter).slice(-1)
      : '',
  inactivityTimeoutUnit:
    inactivityTimeout && String(inactivityTimeout) !== 'Never'
      ? String(inactivityTimeout).slice(-1)
      : '',
  deleteAfterValue:
    deleteAfter && String(deleteAfter) !== 'Never'
      ? String(deleteAfter).slice(0, -1)
      : 'Never',
  inactivityTimeoutValue:
    inactivityTimeout && String(inactivityTimeout) !== 'Never'
      ? String(inactivityTimeout).slice(0, -1)
      : 'Never',
});

  const handleTimeUnitChange = (
    value: string,
    field: 'deleteAfter' | 'inactivityTimeout',
  ) => {
    const keyUnit = field === 'deleteAfter' ? 'deleteAfterUnit' : 'inactivityTimeoutUnit';
    const keyValue = field === 'deleteAfter' ? 'deleteAfterValue' : 'inactivityTimeoutValue';

    const newTimeouts = { ...timeouts, [keyUnit]: value };
    setTimeouts(newTimeouts);

    const finalString =
      newTimeouts[keyValue] === 'Never' || !newTimeouts[keyValue]
        ? 'Never'
        : `${String(newTimeouts[keyValue])}${value}`;
    form.setFieldsValue({
      [field]: finalString,
    });
  };

const handleTimeoutValueChange = (
  value: string | number,
  field: 'deleteAfter' | 'inactivityTimeout',
) => {
  const valStr = String(value);
  const keyUnit = field === 'deleteAfter' ? 'deleteAfterUnit' : 'inactivityTimeoutUnit';
  const keyValue = field === 'deleteAfter' ? 'deleteAfterValue' : 'inactivityTimeoutValue';

  
  const newTimeouts =
    valStr === 'Never'
      ? { ...timeouts, [keyValue]: 'Never', [keyUnit]: '' }
      : { ...timeouts, [keyValue]: valStr };
  setTimeouts(newTimeouts);


  const unit = newTimeouts[keyUnit] ?? '';
  const finalString = valStr === 'Never' ? 'Never' : `${valStr}${unit}`;

  form.setFieldsValue({
    [field]: finalString,
    });
};

const validateTimeout = async (_: unknown, val: string) => {
              
    if (!val || String(val) === 'Never') return;
    if (!/^[0-9]+[mhd]$/.test(String(val))) {
      throw new Error(
        "Select 'Never' or enter a valid inactivity timeout format (e.g., 30m, 2h, 1d)",
      );
    }
  };


  const isTimeUnitDisabled = ( field: 'deleteAfter' | 'inactivityTimeout') => {
    
    if (field === 'deleteAfter') {
      return timeouts.deleteAfterValue === 'Never'
    } else {
      return timeouts.inactivityTimeoutValue === 'Never'
    }
    
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
      width="600px"
    >
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

        {/* DeleteAfter Timeout */}
      <Form.Item
        label="Expiration After"
        name='deleteAfter'
        required= {isTimeUnitDisabled('deleteAfter') ? false : true}        
        {...formItemLayout}
        rules={[
          { required: true },
          { validator: validateTimeout },
        ]}
      >
        <div className="flex gap-4 ">

          <Select defaultValue="Never" onChange={value => handleTimeoutValueChange(value, 'deleteAfter')}>
            {Array.from({ length: 61 }, (_, idx) => {
              const i = idx;
              return (
                <Select.Option key={i} value={i == 0 ? 'Never' : i}>
                  {i == 0 ? 'Never' : i}
                </Select.Option>
              );
            })}
          </Select>

          <Select placeholder="Select time unit" onChange={value => handleTimeUnitChange(value, 'deleteAfter')} disabled={isTimeUnitDisabled('deleteAfter')}>
            {timeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
          <Tooltip title="Time, since the creation, after which instances based on this template are automatically deleted. Users will be preemptively notified via email to take action.">
            <InfoCircleFilled />
          </Tooltip>
        </div>
      </Form.Item>

      {/* Inactivity Timeout */}
      <Form.Item
        label="Max Inactivity"
        name='inactivityTimeout'
        required= {isTimeUnitDisabled('inactivityTimeout') ? false : true}
        {...formItemLayout}
        rules={[
          {validator: validateTimeout},
        ]}

        
      >
        <div className="flex gap-4 items-center">

          <Select value={timeouts.inactivityTimeoutValue} onChange={value => handleTimeoutValueChange(value, 'inactivityTimeout')}>
            
            {Array.from({ length: 61 }, (_, idx) => {
              const i = idx;
              return (
                <Select.Option key={i} value={i == 0 ? 'Never' : i}>
                  {i==0 ? 'Never' : i}
                
                </Select.Option>
              );
            })}
          </Select>

          <Select onChange={value => handleTimeUnitChange(value, 'inactivityTimeout')} disabled={isTimeUnitDisabled('inactivityTimeout')} placeholder="Select time unit">
            {timeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
          <Tooltip title={
              <>
                {"Instances based on this template are stopped / deleted (depending on persistance settings). If they're not accessed within this time (in certain special cases, activity might not be correctly detected, see "}
                <a
                  href="https://github.com/netgroup-polito/CrownLabs/blob/master/operators/pkg/instautoctrl/README.md#instance-inactive-termination-controller"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  here
                </a>
                {") for further technical information."}
              </>
            }>
            <InfoCircleFilled />
          </Tooltip>
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
