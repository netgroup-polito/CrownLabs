import type { FC } from 'react';
import { useState, useContext, useEffect, useCallback, useMemo, useRef } from 'react';
import { Modal, Form, Input, InputNumber, Select, Tooltip, Checkbox, Collapse, theme, Typography, Space, Flex } from 'antd';
import { Button } from 'antd';
import type { CreateTemplateMutation, ImagesQuery } from '../../../generated-types';
import { InfoCircleOutlined, CheckSquareFilled, CloseSquareFilled } from '@ant-design/icons';
import type { RuleObject } from 'antd/es/form';
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
import type { Interval, TemplateForm } from './types';
import {

  formItemLayout,
  getDefaultTemplate,
  getImageNameNoVer,
  internalRegistry,
  useImageLists,
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

const STATUS_ICON_COLORS = {
  on: '#52c41a',
  off: '#c1c1c1ff',
};

const StatusIcon = ({ active }: { active: boolean }) => (
  active
    ? <CheckSquareFilled style={{ color: STATUS_ICON_COLORS.on }} />
    : <CloseSquareFilled style={{ color: STATUS_ICON_COLORS.off }} />
);

const TimeUnitOptions = [
  { label: 'Minutes', value: 'm' },
  { label: 'Hours', value: 'h' },
  { label: 'Days', value: 'd' },
];

const parseTimeoutString = (s?: string) => {
  if (!s || s === 'never') return { value: 0, unit: 'd' }
  const m = String(s).trim().match(/^(\d+)\s*([mhd])$/i)
  if (!m) return { value: 0, unit: 'd' }

  const unitOpt = TimeUnitOptions.find(
    opt => opt.value === m[2].toLowerCase(),
  );

  return { value: Number(m[1]), unit: unitOpt ? unitOpt.value : 'd' }
};

/** Read an optional runtime config variable injected by Helm via configmap. */
const getDefaultTimeout = (name: string): string =>
  (window as unknown as Record<string, unknown>)[name] as string ?? 'never';

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
  const environments = Form.useWatch('environments', form);
  const isPersistent = (environments ?? form.getFieldValue('environments') ?? template?.environments)?.some(env => env?.persistent) ?? false;

  const stopInputRef = useRef<React.ComponentRef<typeof InputNumber>>(null);
  const deleteInactivityInputRef = useRef<React.ComponentRef<typeof InputNumber>>(null);
  const deleteCreationInputRef = useRef<React.ComponentRef<typeof InputNumber>>(null);


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

  
    const { 
      availableImagesVM, 
      availableImagesContainer, 
      projectBaseNameVM,
      projectBaseNameContainer
    } = useImageLists(dataImages?? {} as ImagesQuery);


  // Determine the final image URL
  const parseImage = (envType: EnvironmentType, image: string): string => {
    if (envType === EnvironmentType.VirtualMachine) {
      
      const selectedImage = availableImagesVM.find(
        i => getImageNameNoVer(i.name) === image,
      );

      
      if (selectedImage) {
        return `${internalRegistry}/${projectBaseNameVM}/${selectedImage.name}`;
      }
    } else if (envType === EnvironmentType.Standalone) {
      
      const selectedImage = availableImagesContainer.find(
        i => getImageNameNoVer(i.name) === image,
      );

      if (selectedImage) {
        return `${internalRegistry}/${projectBaseNameContainer}/${selectedImage.name}`;
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
    let nodeSelectorObject: { [key: string]: string } | undefined | null;
    if (nodeSelectorMode === NodeSelectorOptionMap['SelectAnyNode']) {
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
    } else if (nodeSelectorMode === NodeSelectorOptionMap['NodeSelectorDisabled'] && props.template) {
      // If we're in edit mode (props.template exists) and mode is Automatic,
      // set to null to indicate removal from the template
      nodeSelectorObject = null;
    }


    const parsedTemplate = {
      ...template,
      allowPublicExposure: isPublicExposureEnabled,
      description: template.description || template.name,
      cleanup: {
        stopAfterInactivity: timeouts.stopAfterInactivity.value === 0 ? 'never' : `${timeouts.stopAfterInactivity.value}${timeouts.stopAfterInactivity.unit}`,
        deleteAfterInactivity: (!isPersistent || timeouts.deleteAfterInactivity.value === 0) ? 'never' : `${timeouts.deleteAfterInactivity.value}${timeouts.deleteAfterInactivity.unit}`,
        deleteAfterCreation: timeouts.deleteAfterCreation.value === 0 ? 'never' : `${timeouts.deleteAfterCreation.value}${timeouts.deleteAfterCreation.unit}`,
      },
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
        stopAfterInactivity: { value: 0, unit: '' },
        deleteAfterInactivity: { value: 0, unit: '' },
        deleteAfterCreation: { value: 0, unit: '' },
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
    // Blur timeout InputNumbers to commit any typed value before validation.
    // These inputs are not direct Form.Item children, so typed values are
    // synced manually via their onBlur handlers.
    stopInputRef.current?.blur();
    deleteInactivityInputRef.current?.blur();
    deleteCreationInputRef.current?.blur();
    try {
      await form.validateFields();
    } catch (error) {
      console.error('ModalCreateTemplate validation error:', error);
    }
  };
  
  const [timeouts, setTimeouts] = useState(
    {
      stopAfterInactivity: template
        ? { value: parseTimeoutString(template.cleanup?.stopAfterInactivity).value ?? 0, unit: parseTimeoutString(template.cleanup?.stopAfterInactivity).unit }
        : parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_STOP_AFTER_INACTIVITY')),
      deleteAfterInactivity: template
        ? { value: parseTimeoutString(template.cleanup?.deleteAfterInactivity).value ?? 0, unit: parseTimeoutString(template.cleanup?.deleteAfterInactivity).unit }
        : parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_DELETE_AFTER_INACTIVITY')),
      deleteAfterCreation: template
        ? { value: parseTimeoutString(template.cleanup?.deleteAfterCreation).value ?? 0, unit: parseTimeoutString(template.cleanup?.deleteAfterCreation).unit }
        : parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_DELETE_AFTER_CREATION')),
    });

  const {
    data: labelsData,
    loading: loadingLabels,
    error: labelsError,
  } = useNodesLabelsQuery({
    fetchPolicy: 'cache-first',
    skip: !show, // Only fetch when modal is open
  });


  useEffect(() => {
    if (!show) return;

    if (template) {
      const initial = getInitialValues(template);
      form.setFieldsValue(initial);
      setTimeouts({
        stopAfterInactivity: parseTimeoutString(initial.cleanup?.stopAfterInactivity),
        deleteAfterInactivity: parseTimeoutString(initial.cleanup?.deleteAfterInactivity),
        deleteAfterCreation: parseTimeoutString(initial.cleanup?.deleteAfterCreation),
      });
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
          // Only process if labelsData is available
          if (labelsData?.labels) {
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
        }
      } else {
        setNodeSelectorMode(NodeSelectorOptionMap['NodeSelectorDisabled']);
        setSelectedLabels([]);
      }

    } else {
      form.resetFields();
      form.setFieldsValue(getInitialValues(undefined));
      setTimeouts({
        stopAfterInactivity: parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_STOP_AFTER_INACTIVITY')),
        deleteAfterInactivity: parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_DELETE_AFTER_INACTIVITY')),
        deleteAfterCreation: parseTimeoutString(getDefaultTimeout('VITE_APP_DEFAULT_DELETE_AFTER_CREATION')),
      });
      setNodeSelectorMode(NodeSelectorOptionMap['NodeSelectorDisabled']);
      setSelectedLabels([]);
      setIsPublicExposureEnabled(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [template, show, form, labelsData]);

  const NodeSelectorOptionMap: { [key: string]: string } = {
    'NodeSelectorDisabled': 'Automatic',
    'SelectAnyNode': 'Let user choose',
    'FixedSelection': 'Fixed Labels',
  };
  const nodeSelectorTooltips: { [key: string]: string } = {
    'NodeSelectorDisabled': 'No node selection constraints will be applied',
    'SelectAnyNode': 'User can select any node available in the cluster when creating an instance based on this template',
    'FixedSelection': 'Select specific node labels to constrain where instances can run',
  };

  const [nodeSelectorMode, setNodeSelectorMode] = useState<string>(NodeSelectorOptionMap['NodeSelectorDisabled']);
  const [selectedLabels, setSelectedLabels] = useState<string[]>([]);
  const [isPublicExposureEnabled, setIsPublicExposureEnabled] = useState(false);


  const handleSelectorLabelChange = useCallback((values: string[]) => {

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
    setSelectedLabels(filteredValues);
  }, []);

  const handleNodeSelectorModeChange = useCallback((value: string) => {
    setNodeSelectorMode(value);
    if (value === NodeSelectorOptionMap['NodeSelectorDisabled']) {
      setSelectedLabels([]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleTimeoutValueChange = (value: number | null, field: 'stopAfterInactivity' | 'deleteAfterInactivity' | 'deleteAfterCreation') => {
    const numValue = value ? Number(value) : 0;
    const unit = timeouts[field].unit;
    setTimeouts(prevTimeouts => ({
      ...prevTimeouts,
      [field]: { value: numValue, unit },
    }));
    form.setFieldsValue({ cleanup: { [field]: { value: numValue, unit } } });
    if (field === 'stopAfterInactivity' || field === 'deleteAfterCreation') {
      form.validateFields([['cleanup', 'stopAfterInactivity'], ['cleanup', 'deleteAfterCreation']]).catch(() => { });
    } else {
      form.validateFields([['cleanup', field]]).catch(() => { });
    }
  }

  const handleTimeUnitChange = (newUnit: string, field: 'stopAfterInactivity' | 'deleteAfterInactivity' | 'deleteAfterCreation') => {
    const val = timeouts[field].value;
    setTimeouts(prevTimeouts => ({
      ...prevTimeouts,
      [field]: { value: val, unit: newUnit },
    }));
    form.setFieldsValue({ cleanup: { [field]: { value: val, unit: newUnit } } });
    if (field === 'stopAfterInactivity' || field === 'deleteAfterCreation') {
      form.validateFields([['cleanup', 'stopAfterInactivity'], ['cleanup', 'deleteAfterCreation']]).catch(() => { });
    } else {
      form.validateFields([['cleanup', field]]).catch(() => { });
    }
  }

  const handleTimeoutInputSync = (e: React.FocusEvent<HTMLInputElement>, field: 'stopAfterInactivity' | 'deleteAfterInactivity' | 'deleteAfterCreation') => {
    const val = Number(e.target.value) || 0;
    handleTimeoutValueChange(val, field);
  };

  const isTimeUnitDisabled = (field: 'stopAfterInactivity' | 'deleteAfterInactivity' | 'deleteAfterCreation') => {
    return timeouts[field].value === 0;
  };

  const validateTimeout = async (_: RuleObject, _val: { value: number; unit: string } | undefined) => {
    if (!_val || _val.value === undefined || _val.value === 0) {
      return true;
    }

    if (TimeUnitOptions.map(option => option.value).includes(_val.unit) === false) {
      throw new Error("Insert a valid time unit");
    }
    return true;
  };

  const validateTimeoutOrder = async (_: RuleObject, _val: { value: number; unit: string } | undefined, field: 'stopAfterInactivity' | 'deleteAfterInactivity' | 'deleteAfterCreation') => {

    const toMinutes = (t: { value: number; unit: string } | string | undefined) => {
      if (!t) return undefined;
      let obj: { value: number; unit: string };
      if (typeof t === 'string') {
        obj = parseTimeoutString(t);
      } else {
        obj = t as { value: number; unit: string };
      }
      if (obj.value === 0) return Infinity;
      const u = String(obj.unit || '').toLowerCase();
      const mul = u === 'h' ? 60 : u === 'd' ? 1440 : 1;
      return Number(obj.value) * mul;
    };

    const current = form.getFieldValue(['cleanup', field]);
    const stopAfterInactivity = field === 'stopAfterInactivity' ? current : form.getFieldValue(['cleanup', 'stopAfterInactivity']);
    const deleteAfterCreation = field === 'deleteAfterCreation' ? current : form.getFieldValue(['cleanup', 'deleteAfterCreation']);

    if (!stopAfterInactivity || !deleteAfterCreation) return;

    const stopAfterInactivityMin = toMinutes(stopAfterInactivity);
    const deleteAfterCreationMin = toMinutes(deleteAfterCreation);

    if (deleteAfterCreationMin === Infinity) return;

    if (typeof stopAfterInactivityMin !== 'number' || typeof deleteAfterCreationMin !== 'number') return;

    if (!isTimeUnitDisabled('stopAfterInactivity') && stopAfterInactivityMin >= deleteAfterCreationMin) {
      throw new Error('Stop time must be smaller than Expire time');
    }
    return;
  };

  const [infoNumberTemplate, setInfoNumberTemplate] = useState<number>(template?.environments?.length ?? 1);


  // Memoize processed labels to avoid recalculating on every render
  const getNodeLabelsOptions = useMemo(() => {
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
  }, [labelsData, loadingLabels, labelsError]);

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

  const automaticInstanceSavingResource = <>
    <style>{`
      .right-align-error .ant-form-item-explain-error {
        text-align: right;
      }
      .multiline-label .ant-form-item-label > label {
        height: auto !important;
        white-space: normal !important;
        align-items: flex-start !important;
      }
    `}</style>
    <Typography.Paragraph type="secondary" italic className="mb-4">
      Set the value to 0 to disable the corresponding feature
    </Typography.Paragraph>
    <Form.Item
      className="right-align-error multiline-label"
      colon={false}
      label={<div className="flex flex-col text-left"><span>Power off if inactive for:</span><Typography.Text keyboard className="w-max mt-1">Stop</Typography.Text></div>}
      name={['cleanup', 'stopAfterInactivity']}
      validateTrigger="onChange"
      rules={[{ validator: validateTimeout }, { validator: (rule, value) => validateTimeoutOrder(rule, value, 'stopAfterInactivity') }]}
      {...formItemLayout}>

      <div className="flex flex-1 w-full items-center justify-between">
        <Tooltip title={<><p>Instances based on this template are stopped / deleted (based on their persistency) if they're not accessed within this time (in certain special cases, activity might not be correctly detected, see <a href='https://github.com/netgroup-polito/CrownLabs/blob/master/operators/pkg/instautoctrl/README.md#instance-inactive-termination-controller'>here</a> for further technical information).</p> <p> <b>Set 0 to disable the feature.</b></p></>}>
          <InfoCircleOutlined className='ml-2' />
        </Tooltip>
        <div className="flex gap-4 items-center">
          <InputNumber
            ref={stopInputRef}
            onChange={value => handleTimeoutValueChange(value, 'stopAfterInactivity')}
            onBlur={e => handleTimeoutInputSync(e, 'stopAfterInactivity')}
            min={0}
            defaultValue={timeouts.stopAfterInactivity.value}
          />

          <Select
            style={{ width: 130 }}
            onChange={value => handleTimeUnitChange(value, 'stopAfterInactivity')}
            disabled={isTimeUnitDisabled('stopAfterInactivity')}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
            defaultValue={parseTimeoutString(template?.cleanup?.stopAfterInactivity).unit}

          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </div>
    </Form.Item>

    <Form.Item
      className="right-align-error multiline-label"
      colon={false}
      label={<div className="flex flex-col text-left"><span>Delete if powered off for:</span><Typography.Text keyboard className="w-max mt-1">Delete</Typography.Text></div>}
      name={['cleanup', 'deleteAfterInactivity']}
      validateTrigger="onChange"
      rules={[{ validator: validateTimeout }]}
      {...formItemLayout}>

      <div className="flex flex-1 w-full items-center justify-between">
        <Tooltip title={<><p>Instances based on this template are deleted if they're not powered on within this time.</p> {isPersistent ? <b>Set 0 to disable the feature.</b> : <b>Not applicable for non-persistent instances.</b>}</>}>
          <InfoCircleOutlined className='ml-2' />
        </Tooltip>
        <div className="flex gap-4 items-center">
          <InputNumber
            ref={deleteInactivityInputRef}
            onChange={value => handleTimeoutValueChange(value, 'deleteAfterInactivity')}
            onBlur={e => handleTimeoutInputSync(e, 'deleteAfterInactivity')}
            min={0}
            value={!isPersistent ? 0 : timeouts.deleteAfterInactivity.value}
            disabled={!isPersistent}
          />

          <Select
            style={{ width: 130 }}
            onChange={value => handleTimeUnitChange(value, 'deleteAfterInactivity')}
            disabled={!isPersistent || isTimeUnitDisabled('deleteAfterInactivity')}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
            value={!isPersistent ? 'd' : timeouts.deleteAfterInactivity.unit}
          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </div>
    </Form.Item>
    <Form.Item
      className="right-align-error multiline-label"
      colon={false}
      label={<div className="flex flex-col text-left"><span>Delete regardless of activity after:</span><Typography.Text keyboard className="w-max mt-1">Expire</Typography.Text></div>}
      name={['cleanup', 'deleteAfterCreation']}
      validateTrigger="onChange"
      rules={[{ validator: validateTimeout }]}
      {...formItemLayout}>

      <div className="flex flex-1 w-full items-center justify-between">
        <Tooltip title={<><p>Time, since the creation, after which instances based on this template are automatically deleted. Users will be preemptively alerted through email to take actions.</p> <p><b>Set 0 to disable the feature.</b></p></>}>

          <InfoCircleOutlined className='ml-2' />
        </Tooltip>
        <div className="flex gap-4 items-center">
          <InputNumber
            ref={deleteCreationInputRef}
            onChange={value => handleTimeoutValueChange(value, 'deleteAfterCreation')}
            onBlur={e => handleTimeoutInputSync(e, 'deleteAfterCreation')}
            min={0}
            defaultValue={timeouts.deleteAfterCreation.value}
          />

          <Select
            style={{ width: 130 }}
            onChange={value => handleTimeUnitChange(value, 'deleteAfterCreation')}
            disabled={isTimeUnitDisabled('deleteAfterCreation')}
            placeholder="Select Time unit"
            getPopupContainer={trigger => trigger.parentElement || document.body}
            defaultValue={parseTimeoutString(template?.cleanup?.deleteAfterCreation).unit}
          >
            {TimeUnitOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                {option.label}
              </Select.Option>
            ))}
          </Select>
        </div>
      </div>
    </Form.Item>
  </>

  const environmentListForm = <>
  <EnvironmentList
          availableImagesVM={availableImagesVM}
          availableImagesContainer={availableImagesContainer}
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
    {/* TODO: public exporsure, nodeselector, template description */}
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

    <Flex justify='space-around' className="mb-0 gap-2"  {...formItemLayout} align="center">
      <Space direction='vertical' style={{ width: "50%" }}>
        <Typography.Paragraph className="mb-0">Server Type: <Tooltip title="Allow instances based on this template to be scheduled on specific nodes"><InfoCircleOutlined className='ml-1' /></Tooltip></Typography.Paragraph>
        <Select
          style={{ width: "100%" }}
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
      <Space direction='vertical' style={{ width: "50%" }}>
        {nodeSelectorMode === NodeSelectorOptionMap['FixedSelection'] && (<>
          <Typography.Paragraph className="mb-0">Labels: <Tooltip title={<span>Select on which node types instances based on this template can be scheduled. This option is enabled only if <strong>Fixed</strong> is selected. For the same tag, only one value can be selected (e.g. nodeSize=big and nodeSize=small cannot be selected simultaneously).</span>}><InfoCircleOutlined className='ml-1' /></Tooltip></Typography.Paragraph>
          <Select
            disabled={nodeSelectorMode !== NodeSelectorOptionMap['FixedSelection']}
            style={{ width: "100%" }}
            mode="multiple"
            placeholder="Select"
            onChange={handleSelectorLabelChange}
            options={getNodeLabelsOptions}
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

        <Collapse size="small" bordered={false} ghost accordion items={[
          {
            key: '1',
            label: <Typography.Text strong>Virtual Machines / Containers</Typography.Text>,
            children: environmentListForm,
            style: panelStyle,
            forceRender: true,
            extra: <Text keyboard>{infoNumberTemplate ? infoNumberTemplate == 1 ? '1 environment' : `${infoNumberTemplate} environments` : 'No environments'}</Text>
          },
          {
            key: '2',
            label: <Typography.Text strong>Automatic Clean-up</Typography.Text>,
            children: automaticInstanceSavingResource,
            style: panelStyle,
            forceRender: true,
            extra: <><Text keyboard>Stop <StatusIcon active={!isTimeUnitDisabled('stopAfterInactivity')} /></Text> <Text keyboard delete={!isPersistent}>Delete <StatusIcon active={isPersistent && !isTimeUnitDisabled('deleteAfterInactivity')} /></Text> <Text keyboard>Expire <StatusIcon active={!isTimeUnitDisabled('deleteAfterCreation')} /></Text></>
          },
          {
            key: '3',
            label: <Typography.Text strong>Advanced Features</Typography.Text>,
            children: advancedFeaturesForm,
            forceRender: true,
            style: panelStyle,
            extra: <><Text keyboard>Exposure <StatusIcon active={isPublicExposureEnabled} /></Text> <Text keyboard>Node Selector <StatusIcon active={nodeSelectorMode !== NodeSelectorOptionMap['NodeSelectorDisabled']} /></Text></>
          },
        ]} defaultActiveKey={['1']} />


        <div className="flex justify-end gap-2">
          <Button type="default" onClick={() => closehandler()}>
            Cancel
          </Button>

          <Form.Item shouldUpdate>
            {() => {
              const fieldsError = form.getFieldsError();
              const hasErrors = fieldsError.some(
                ({ errors }) => errors.length > 0,
              );

              // Check required fields
              const templateName = form.getFieldValue('name');
              const environments = form.getFieldValue('environments') as TemplateForm['environments'];

              const hasTemplateName = templateName && templateName.trim() !== '';

              // ALL environments must have all required fields filled
              const hasValidEnvironments = environments && environments.length > 0 &&
                environments.every(env =>
                  env.name && env.name.trim() !== '' &&
                  env.environmentType &&
                  env.image && env.image.trim() !== ''
                );

              // Node selector validation
              const nodeSelectorValid = nodeSelectorMode !== NodeSelectorOptionMap['FixedSelection'] ||
                selectedLabels.length > 0;

              const isDisabled = hasErrors || !hasTemplateName || !hasValidEnvironments || !nodeSelectorValid;

              return (
                <span>
                  <Button htmlType="submit" type="primary" disabled={isDisabled}>
                    {!loading && (template ? 'Modify' : 'Create')}
                  </Button>
                </span>
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
