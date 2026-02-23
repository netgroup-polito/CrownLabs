import type { FC } from 'react';
import { useState, useContext, useEffect, useCallback } from 'react';
import { Modal, Form, Input, InputNumber, Select, Tooltip, Checkbox, Collapse, theme, Typography, Space, Flex } from 'antd';
import { Button } from 'antd';
import type { CreateTemplateMutation } from '../../../generated-types';
import { InfoCircleOutlined } from '@ant-design/icons';
import type { RuleObject } from 'antd/es/form';
import type { CollapseProps } from 'antd';
import {
  useNodesLabelsQuery,
} from '../../../generated-types';
import {
  EnvironmentType,
  useWorkspaceTemplatesQuery,
  useImagesQuery,
  useWorkspaceSharedVolumesQuery,
} from '../../../generated-types';
import type { ApolloError, FetchResult } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { makeGuiSharedVolume } from '../../../utilsLogic';
import { cleanupLabels, type SharedVolume } from '../../../utils';
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

const { Text } = Typography;


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

const TimeUnitOptions = [
    { label: 'Minutes', value: 'm' },
    { label: 'Hours', value: 'h' },
    { label: 'Days', value: 'd' },
  ];

const parseTimeoutString = (s?: string) => {
    if (!s || s === 'never') return { value: 0, unit: '' }
    const m = String(s).trim().match(/^(\d+)\s*([mhd])$/i)
    if (!m) return { value: 0, unit: '' }
    
    const unitOpt = TimeUnitOptions.find(
      opt => opt.value === m[2].toLowerCase(),
    );

    return { value: Number(m[1]), unit: unitOpt ? unitOpt.value : ''}
  };

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
    if (template) { // we are editing an existing template, not creating a new one
      return;
    }
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
    form.resetFields();
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
    let nodeSelectorObject: { [key: string]: string } | undefined;
    if (nodeSelectorMode === NodeSelectorOptionMap['AnyNode']) {
      nodeSelectorObject = {};
    } else if (nodeSelectorMode === NodeSelectorOptionMap['FixedSelection'] && selectedLabels.length > 0) {
      nodeSelectorObject = selectedLabels.reduce((acc, jsonStr) => {
        try {
          const labelPair = JSON.parse(jsonStr);
          return { ...acc, ...labelPair };
        } catch (e) {
          console.error('Error parsing label:', e);
          return acc;
        }
      }, {});
    }

    
    const parsedTemplate = {
      ...template,
      description: template.description || template.name,
      inactivityTimeout: timeouts.inactivityTimeout.value === 0 ? 'never' : `${timeouts.inactivityTimeout.value}${timeouts.inactivityTimeout.unit}`,
      deleteAfter: timeouts.deleteAfter.value === 0 ? 'never' : `${timeouts.deleteAfter.value}${timeouts.deleteAfter.unit}`,
      environments: template.environments.map(env => ({
        ...env,
        image: parseImage(env.environmentType, env.image),
      })),
      ...(nodeSelectorObject !== undefined && { nodeSelector: nodeSelectorObject }),
    };
    try {
      setShow(false);
      await submitHandler(parsedTemplate);
      
      form.resetFields();
      setTimeouts({
        inactivityTimeout: { value: 0, unit: '' },
        deleteAfter: { value: 0, unit: '' },
      });
      setNodeSelectorMode('Disabled');
      setSelectedLabels([]);
    } catch (error) {
      console.error('ModalCreateTemplate submitHandler error:', error);
      apolloErrorCatcher(error as ApolloError);
    }
  };

  const getInitialValues = useCallback((template?: TemplateForm) => {
    if (template) return template;

    return getDefaultTemplate({
      cpu: cpuInterval,
      ram: ramInterval,
      disk: diskInterval,
    });
  }, [cpuInterval, ramInterval, diskInterval]);

  const handleFormSubmit = async () => {
    try {
      await form.validateFields();
    } catch (error) {
      console.error('ModalCreateTemplate validation error:', error);
    }
  };

  const [timeouts, setTimeouts] = useState(
    {
    inactivityTimeout: { value: parseTimeoutString(template?.inactivityTimeout).value ?? 0, unit: parseTimeoutString(template?.inactivityTimeout).unit ?? '' },
    deleteAfter: { value: parseTimeoutString(template?.deleteAfter).value ?? 0, unit: parseTimeoutString(template?.deleteAfter).unit ?? '' },
  });

  useEffect(() => {
  if (!show) return;

  if (template) {
    const initial = getInitialValues(template);
    form.setFieldsValue(initial);
    setTimeouts({
      inactivityTimeout: parseTimeoutString(initial.inactivityTimeout),
      deleteAfter: parseTimeoutString(initial.deleteAfter),
    });
    setAutomaticStoppingEnabled(
      (initial.inactivityTimeout) !== 'never' ||
        (initial.deleteAfter) !== 'never',
    );
    setIsPublicExposureEnabled(initial.allowPublicExposure ?? false);
      // Set node selector mode and labels based on template
      if (template.nodeSelector) {
        if (Object.keys(template.nodeSelector).length === 0) {
          setNodeSelectorMode(NodeSelectorOptionMap['SelectAnyNode']);
          setSelectedLabels([]);
        } else {
          setNodeSelectorMode(NodeSelectorOptionMap['FixedSelection']);
          // Convert nodeSelector object to JSON values matching the Select options
          // Map camelCase keys back to original format using available labels
          const jsonValues = Object.entries(template.nodeSelector)
            .map(([key, value]) => {
              const originalLabel = findOriginalLabelKey(key, value as string);
              if (originalLabel) {
                return JSON.stringify({ [originalLabel.key]: originalLabel.value });
              }
              // Fallback to the key as-is if we can't find a match
              console.warn('Could not find original label for:', key, value);
              return JSON.stringify({ [key]: value });
            });
          setSelectedLabels(jsonValues);
        }
      } else {
        setNodeSelectorMode(NodeSelectorOptionMap['NodeSelectorDisabled']);
        setSelectedLabels([]);
      }

  } else {
    form.resetFields();
    form.setFieldsValue(getInitialValues(undefined));
    setTimeouts({
      inactivityTimeout: { value: 0, unit: '' },
      deleteAfter: { value: 0, unit: '' },
    });
    setAutomaticStoppingEnabled(false);
    setNodeSelectorMode(NodeSelectorOptionMap['NodeSelectorDisabled']);
    setSelectedLabels([]);
    setIsPublicExposureEnabled(false);
  }
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, [template, show, form, getInitialValues]);

  const NodeSelectorOptionMap: { [key: string]: string } = {
    'NodeSelectorDisabled': 'Disabled',
    'SelectAnyNode': 'Any node',
    'FixedSelection': 'Fixed',
  };
  const nodeSelectorTooltips: { [key: string]: string } = {
    'NodeSelectorDisabled': 'No node selection constraints will be applied',
    'SelectAnyNode': 'User can select any node available in the cluster when creating an instance based on this template',
    'FixedSelection': 'Select specific node labels to constrain where instances can run',
  };

  const [automaticStoppingEnabled, setAutomaticStoppingEnabled] = useState(false);
  const [nodeSelectorMode, setNodeSelectorMode] = useState<string>(NodeSelectorOptionMap['NodeSelectorDisabled']);
  const [selectedLabels, setSelectedLabels] = useState<string[]>([]);
  const [isPublicExposureEnabled, setIsPublicExposureEnabled] = useState(false);

   const getNodeLabelsOptions = () => {
    if (loadingLabels || labelsError) {
      return [
        {
          value: 'error',
          label: loadingLabels ? 'Loading labels...' : 'Error loading labels',
          disabled: true,
        },
      ];
    }
    
    return labelsData?.labels?.map(({ key, value }) => ({
      value: JSON.stringify({ [key]: value }),
      label: `${cleanupLabels(key)}=${value}`,
    })) ?? [];
  };

  
const handleSelectorLabelChange = useCallback((values: string[]) => {
  console.log('Selected label values:', values);
  
  // Filter out duplicate keys - keep only the last selected value for each key
  const seenKeys = new Map<string, string>();
  const filteredValues: string[] = [];
  
  // Process in reverse to keep the most recent selection for each key
  for (let i = values.length - 1; i >= 0; i--) {
    try {
      const labelPair = JSON.parse(values[i]);
      const key = Object.keys(labelPair)[0];
      
      if (!seenKeys.has(key)) {
        seenKeys.set(key, values[i]);
        filteredValues.unshift(values[i]); // Add to beginning to maintain order
      }
    } catch (e) {
      console.error('Error parsing label:', e);
    }
  }
  console.log('Filtered label values (duplicates removed):', filteredValues);
  setSelectedLabels(filteredValues);
}, []);

const handleNodeSelectorModeChange = useCallback((value: string) => {
  setNodeSelectorMode(value);
  if (value === NodeSelectorOptionMap['NodeSelectorDisabled']) {
    setSelectedLabels([]);
  }
}, []);

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

  const validateTimeoutOrder = async (_: RuleObject , _val: { value: number; unit: string } | undefined, field: 'inactivityTimeout' | 'deleteAfter') => {
   
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

  const [infoNumberTemplate, setInfoNumberTemplate] = useState<number>(template?.environments?.length ?? 1 );

  const {
      data: labelsData,
      loading: loadingLabels,
      error: labelsError,
    } = useNodesLabelsQuery({ fetchPolicy: 'no-cache' });

  // Helper function to map camelCase keys back to original format
  const findOriginalLabelKey = (camelCaseKey: string, value: string): { key: string; value: string } | null => {
    if (!labelsData?.labels) return null;
    
    // First, try exact match (in case the key wasn't camelCased)
    const exactMatch = labelsData.labels.find(
      label => label.key === camelCaseKey && label.value === value
    );
    if (exactMatch) return exactMatch;
    
    // Otherwise, find by matching cleaned version
    const cleanedCamelCase = cleanupLabels(camelCaseKey);
    const match = labelsData.labels.find(
      label => cleanupLabels(label.key) === cleanedCamelCase && label.value === value
    );
    
    return match || null;
  };

    const handleEnablingCleanUp = (enabled: boolean) => {
      setAutomaticStoppingEnabled(enabled);
      if (!enabled) {
        // If disabling, reset timeouts to 0
        setTimeouts({
          inactivityTimeout: { value: 0, unit: '' },
          deleteAfter: { value: 0, unit: '' },
        });
        form.setFieldValue('inactivityTimeout', { value: 0, unit: '' });
        form.setFieldValue('deleteAfter', { value: 0, unit: '' });
        form.validateFields(['inactivityTimeout', 'deleteAfter']).catch(() => {});
      }
    };

  const automaticInstanceSavingResource = <>
  <Checkbox className="mb-4" checked={automaticStoppingEnabled} onChange={e => handleEnablingCleanUp(e.target.checked)}>Enable automatic clean-up</Checkbox>
        
      <Form.Item
        label="Max Inactivity"
        name="inactivityTimeout"
        required={isTimeUnitDisabled('inactivityTimeout') ? false : true}
        validateTrigger="onChange"
        rules={[{ validator: validateTimeout }, { validator: (rule, value) => validateTimeoutOrder(rule, value, 'inactivityTimeout') }]}
        {...formItemLayout}> 
        
        <div className="flex gap-4 items-center">
          <Tooltip title={<><p>Instances based on this template are stopped / deleted (based on their persistency) if they're not accessed within this time (in certain special cases, activity might not be correctly detected, see <a href='https://github.com/netgroup-polito/CrownLabs/blob/master/operators/pkg/instautoctrl/README.md#instance-inactive-termination-controller'>here</a> for further technical information).</p> <p> <b>Set 0 to disable the feature.</b></p></>}>
            <InfoCircleOutlined className='ml-2'/>
          </Tooltip>
          <InputNumber
            onChange={value => handleTimeoutValueChange(value, 'inactivityTimeout')}
            min={0}
            max={60}
            defaultValue={timeouts.inactivityTimeout.value }
            disabled={!automaticStoppingEnabled}
          >
          </InputNumber>

          <Select
            onChange={value => handleTimeUnitChange(value, 'inactivityTimeout')}
            disabled={isTimeUnitDisabled('inactivityTimeout') || !automaticStoppingEnabled}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
            defaultValue={parseTimeoutString(template?.inactivityTimeout).unit}
            
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
        label="Max Lifetime"
        name="deleteAfter"
        required={isTimeUnitDisabled('deleteAfter') ? false : true}
        validateTrigger="onChange"
        rules={[{ validator: validateTimeout }]}
        {...formItemLayout}> 
        
        <div className="flex gap-4 items-center">
          <Tooltip title={<><p>Time, since the creation, after which instances based on this template are automatically deleted. Users will be preemptively alerted through email to take actions.</p> <p><b>Set 0 to disable the feature.</b></p></>}>
          
            <InfoCircleOutlined className='ml-2'/>
          </Tooltip>
          <InputNumber
            onChange={value => handleTimeoutValueChange(value, 'deleteAfter')}
            min={0}
            max={60}
            defaultValue={timeouts.deleteAfter.value}
            disabled={!automaticStoppingEnabled}
          >
          </InputNumber>

          <Select
            onChange={value => handleTimeUnitChange(value, 'deleteAfter')}
            disabled={isTimeUnitDisabled('deleteAfter') || !automaticStoppingEnabled}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
            defaultValue={parseTimeoutString(template?.deleteAfter).unit}
          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </Form.Item>
      </>

  const environmentListForm = <>
  <EnvironmentList
          availableImages={availableImages}
          resources={{
            cpu: cpuInterval,
            ram: ramInterval,
            disk: diskInterval,
          }}
          sharedVolumes={sharedVolumes}
          setInfoNumberTemplate={setInfoNumberTemplate}
          isPersonal={isPersonal === undefined ? false : isPersonal}
        /></>
  



  const advancedFeaturesForm = <>
    {/* TODO: public exporsure, nodeselector, template description */ }
    <Form.Item
      name="description"
      className="mb-4"
      required={false}
      label="Description"
      {...formItemLayout}
      >
    <Input.TextArea
      rows={2}
      placeholder="Insert template description"
      maxLength={250}
    />
    </Form.Item>
          <Form.Item
            name="allowPublicExposure"
            valuePropName="checked"
            className="gap-6 ">
              <Checkbox onChange={(e) => setIsPublicExposureEnabled(e.target.checked)} className='ml-4'>
                Port Exposure / Port Forwarding{' '}
                <Tooltip title="Allow instances based on this template to be publicly accessible via Public IP">
                  <InfoCircleOutlined />
                </Tooltip>
              </Checkbox>
          </Form.Item>
        
   <Flex justify='space-around' className="mb-0 gap-2" {...formItemLayout}  align="center">
    <Space direction='vertical' style={{width:"50%"}}>
      <Typography.Paragraph className="mb-0">Node Selector: <Tooltip title="Allow instances based on this template to be scheduled on specific nodes"><InfoCircleOutlined className='ml-1' /></Tooltip></Typography.Paragraph>
      <Select 
        style={{width:"100%"}} 
        value={nodeSelectorMode}
        onChange={handleNodeSelectorModeChange}

      >
        {NodeSelectorOptionMap && Object.entries(NodeSelectorOptionMap).map(([key, label]) => (
          <Select.Option key={label} value={label}>
            <Tooltip title={nodeSelectorTooltips[key]} placement="left">
              <span>{label}</span>
            </Tooltip>
          </Select.Option>
        ))}
      </Select>
    </Space>
    <Space direction='vertical'  style={{width:"50%"}}>
       {nodeSelectorMode === NodeSelectorOptionMap['FixedSelection'] && (<>
      <Typography.Paragraph className="mb-0">Labels: <Tooltip title={<span>Select on which node types instances based on this template can be scheduled. This option is enabled only if <strong>Fixed</strong> is selected. For the same tag, only one value can be selected (e.g. nodeSize=big and nodeSize=small cannot be selected simultaneously).</span>}><InfoCircleOutlined className='ml-1' /></Tooltip></Typography.Paragraph>
       <Select
          disabled={nodeSelectorMode !== NodeSelectorOptionMap['FixedSelection']}
          style={{width:"100%"}}
          mode="multiple"
          placeholder="Select"
          onChange={handleSelectorLabelChange}
          options={getNodeLabelsOptions()}
          value={selectedLabels}
          status={nodeSelectorMode === NodeSelectorOptionMap['FixedSelection'] && selectedLabels.length === 0 ? 'error' : undefined}
        />
      </>)}
    </Space>
    </Flex>

  </>


  const { token } = theme.useToken();
  const panelStyle: React.CSSProperties = {
    marginBottom: 10,
    background: `${token.colorFillAlter}`,
    borderRadius: token.borderRadiusLG,
    border: `1px solid ${token.colorBorderSecondary}`,
    padding: '0px 10px',
    
  };



  const collapseFormItems: CollapseProps['items'] = [
  {
    key: '1',
    label: <Typography.Text strong>Automatic Clean-up</Typography.Text>,
    children: automaticInstanceSavingResource,
    style: panelStyle,
    extra: <><Text keyboard>{automaticStoppingEnabled && !isTimeUnitDisabled('inactivityTimeout') ? 'Inactivity ON' : 'Inactivity OFF'}</Text> <Text keyboard>{automaticStoppingEnabled && !isTimeUnitDisabled('deleteAfter') ? 'Expiration ON' : 'Expiration OFF'}</Text></>
  },
  {
    key: '2',
    label: <Typography.Text strong>Virtual Machines / Containers</Typography.Text>,
    children: environmentListForm,
    style: panelStyle,
    extra: <Text keyboard>{infoNumberTemplate ? infoNumberTemplate == 1 ? '1 environment set' : `${infoNumberTemplate} environments set` : 'No environments set'}</Text>
  },
  {
    key: '3',
    label: <Typography.Text strong>Advanced Features</Typography.Text>,
    children: advancedFeaturesForm,
    style: panelStyle,
    extra: <><Text keyboard>{isPublicExposureEnabled ? 'Exposure ON' : 'Exposure OFF'}</Text> <Text keyboard>{nodeSelectorMode !== NodeSelectorOptionMap['Disabled'] ? 'Node Selector ON' : 'Node Selector OFF'}</Text></>
  },
];

  return (
    
    <Modal
    
      destroyOnHidden={true}
      styles={{ body: { paddingBottom: '5px' } }}
      centered
      footer={null}
      title={template ? 'Modify template' : 'Create a new template'}
      open={show}
      onCancel={closehandler}
      width="620px">
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
        
          <Collapse size="small" bordered={false} ghost accordion items={collapseFormItems} defaultActiveKey={['2']}  />
        

        <div className="flex justify-end gap-2">
          <Button type="default" onClick={() => closehandler()}>
            Cancel
          </Button>

          <Form.Item shouldUpdate>
            {() => {
              const fieldsError = form.getFieldsError();
              const hasErrors = fieldsError.some(
                ({ errors }) => errors.length > 0,
              ) || (nodeSelectorMode === NodeSelectorOptionMap['FixedSelection'] && selectedLabels.length === 0)

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
