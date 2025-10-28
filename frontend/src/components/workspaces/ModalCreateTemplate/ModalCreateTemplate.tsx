import type { FC } from 'react';
import { useState, useEffect, useContext, useCallback } from 'react';
import {
  Modal,
  Slider,
  Form,
  Input,
  Checkbox,
  Tooltip,
  AutoComplete,
} from 'antd';
import { Button } from 'antd';
import type {
  CreateTemplateMutation,
  SharedVolumeMountsListItem,
} from '../../../generated-types';
import {
  EnvironmentType,
  useWorkspaceTemplatesQuery,
} from '../../../generated-types';
import type { FetchResult } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import ShVolFormItem, { type ShVolFormItemValue } from './ShVolFormItem';

const alternativeHandle = { border: 'solid 2px #1c7afdd8' };

export type Image = {
  name: string;
  vmorcontainer: Array<VmOrContainer>;
  registry: string;
};

type VmOrContainer = EnvironmentType.VirtualMachine | EnvironmentType.Container;

type Environment = {
  name: string;
  image: string;
  registry?: string;
  vmorcontainer?: VmOrContainer;
  persistent: boolean;
  mountMyDrive: boolean;
  gui: boolean;
  cpu: number;
  ram: number;
  disk: number;
  sharedVolumeMountInfos?: SharedVolumeMountsListItem[];
}

type Template = {
  name?: string;
  environmentList: Environment[];
}

type Interval = {
  max: number;
  min: number;
};

type Valid = {
  name: { status: string; help?: string };
  image: { status: string; help?: string };
};
export interface IModalCreateTemplateProps {
  workspaceNamespace: string;
  template?: Template;
  images: Array<Image>;
  cpuInterval: Interval;
  ramInterval: Interval;
  diskInterval: Interval;
  show: boolean;
  setShow: (status: boolean) => void;
  submitHandler: (
    t: Template,
  ) => Promise<
    FetchResult<
      CreateTemplateMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  loading: boolean;
}

const getImageNoVer = (image: string) =>
  image.split(':').length === 2 ? image.split(':')[0] : image;

const isEmptyOrSpaces = (str: string) => !str || str.match(/^ *$/);

const ModalCreateTemplate: FC<IModalCreateTemplateProps> = ({ ...props }) => {
  const {
    show,
    setShow,
    cpuInterval,
    ramInterval,
    diskInterval,
    images,
    template,
    submitHandler,
    loading,
    workspaceNamespace,
  } = props;

  const imagesNoVersion = images.map(x => getImageNoVer(x.name));

  const [buttonDisabled, setButtonDisabled] = useState(true);

  const [formTemplate, setFormTemplate] = useState<Template>({
    name: template && template?.name,
    environmentList: template?.environmentList || [{
      name: 'main',
      image: '',
      registry: '',
      vmorcontainer: EnvironmentType.Container,
      persistent: false,
      mountMyDrive: true,
      gui: true,
      cpu: cpuInterval.min,
      ram: ramInterval.min,
      disk: diskInterval.min,
      sharedVolumeMountInfos: [],
    }],
  });

  const [valid, setValid] = useState<Valid>({
    name: { status: 'success', help: undefined },
    image: { status: 'success', help: undefined },
  });

  const [imagesSearchOptions, setImagesSearchOptions] = useState<Record<number, string[]>>({});

  const addEnvironment = () => {
    setFormTemplate(old => ({
      ...old,
      environmentList: [
        ...old.environmentList,
        {
          name: `env-${old.environmentList.length + 1}`,
          image: '',
          registry: '',
          vmorcontainer: EnvironmentType.Container,
          persistent: false,
          mountMyDrive: true,
          gui: true,
          cpu: cpuInterval.min,
          ram: ramInterval.min,
          disk: diskInterval.min,
          sharedVolumeMountInfos: [],
        }
      ]
    }));
  };

  const removeEnvironment = (index: number) => {
    if (formTemplate.environmentList.length > 1) {
      setFormTemplate(old => ({
        ...old,
        environmentList: old.environmentList.filter((_, i) => i !== index)
      }));
    }
  };

  const updateEnvironment = (index: number, updates: Partial<Environment>) => {
    setFormTemplate(old => ({
      ...old,
      environmentList: old.environmentList.map((env, i) =>
        i === index ? { ...env, ...updates } : env
      )
    }));
  };


  const validateEnvironments = useCallback(() => {
    const errors: string[] = [];

    formTemplate.environmentList.forEach((env, index) => {
      if (!env.name || env.name.trim() === '') {
        errors.push(`Environment ${index + 1}: Name is required`);
      }
      if (!env.image || env.image.trim() === '') {
        errors.push(`Environment ${index + 1}: Image is required`);
      }
    });

    // Check for duplicate environment names
    const names = formTemplate.environmentList.map(env => env.name);
    const duplicates = names.filter((name, index) => names.indexOf(name) !== index);
    if (duplicates.length > 0) {
      errors.push('Environment names must be unique');
    }

    return errors;
  }, [formTemplate.environmentList]);

  const hasChanges = useCallback(() => {
    if (!template) return true;

    if (template.name !== formTemplate.name) return true;

    if (template.environmentList.length !== formTemplate.environmentList.length) return true;

    return formTemplate.environmentList.some((env, index) => {
      const originalEnv = template.environmentList[index];
      if (!originalEnv) return true;

      return (
        originalEnv.name !== env.name ||
        originalEnv.image !== env.image ||
        originalEnv.vmorcontainer !== env.vmorcontainer ||
        originalEnv.gui !== env.gui ||
        originalEnv.persistent !== env.persistent ||
        originalEnv.cpu !== env.cpu ||
        originalEnv.ram !== env.ram ||
        originalEnv.disk !== env.disk ||
        JSON.stringify(originalEnv.sharedVolumeMountInfos) !==
        JSON.stringify(env.sharedVolumeMountInfos)
      );
    });
  }, [template, formTemplate.name, formTemplate.environmentList]);

  useEffect(() => {
    const envErrors = validateEnvironments();
    const hasValidTemplate = formTemplate.name &&
      formTemplate.environmentList.length > 0 &&
      envErrors.length === 0;

    const changesDetected = hasChanges();
    if (hasValidTemplate && valid.name.status === 'success' && changesDetected) {
      setButtonDisabled(false);
    } else {
      setButtonDisabled(true);
    }
  }, [formTemplate, template, valid.name.status, validateEnvironments, hasChanges]);

  const nameValidator = () => {
    if (formTemplate.name === '' || formTemplate.name === undefined) {
      setValid(old => {
        return {
          ...old,
          name: { status: 'error', help: 'Please insert template name' },
        };
      });
    } else if (
      !errorFetchTemplates &&
      !loadingFetchTemplates &&
      dataFetchTemplates?.templateList?.templates
        ?.map(t => t?.spec?.prettyName)
        .includes(formTemplate.name.trim())
    ) {
      setValid(old => {
        return {
          ...old,
          name: {
            status: 'error',
            help: 'This name has already been used in this workspace',
          },
        };
      });
    } else {
      setValid(old => {
        return {
          ...old,
          name: { status: 'success', help: undefined },
        };
      });
    }
  };

  const imageValidator = () => {
    const hasEmptyImages = formTemplate.environmentList.some(env =>
      isEmptyOrSpaces(env.image)
    );
    if (hasEmptyImages) {
      setValid(old => ({
        ...old,
        image: { status: 'error', help: 'Insert an image for each environment' },
      }));
    } else {
      setValid(old => ({
        ...old,
        image: { status: 'success', help: undefined },
      }));
    }
  }

  const [form] = Form.useForm();

  const fullLayout = {
    wrapperCol: { offset: 0, span: 24 },
  };

  const closehandler = () => {
    setShow(false);
  };

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const {
    data: dataFetchTemplates,
    error: errorFetchTemplates,
    loading: loadingFetchTemplates,
    refetch: refetchTemplates,
  } = useWorkspaceTemplatesQuery({
    onError: apolloErrorCatcher,
    variables: { workspaceNamespace },
  });

  const onSubmit = () => {
    const shvolMounts: ShVolFormItemValue[] = form.getFieldValue('shvolss');
    const processedEnvironmentList = formTemplate.environmentList.map(env => {
      const sharedVolumeMountInfos: SharedVolumeMountsListItem[] =
        shvolMounts.map(obj => ({
          sharedVolume: {
            namespace: obj.shvol.split('/')[0],
            name: obj.shvol.split('/')[1],
          },
          mountPath: obj.mountpath,
          readOnly: Boolean(obj.readonly),
        }));

      return {
        ...env,
        image: images.find(i => getImageNoVer(i.name) === env.image)?.name ?? env.image,
        sharedVolumeMountInfos: env.sharedVolumeMountInfos || sharedVolumeMountInfos,
      };
    });


    submitHandler({
      ...formTemplate,
      environmentList: processedEnvironmentList,
    })
      .then(() => {
        setShow(false);
        setFormTemplate(old => ({
          ...old,
          name: undefined,
          environmentList: [{
            name: 'main',
            image: '',
            registry: '',
            vmorcontainer: EnvironmentType.Container,
            persistent: false,
            mountMyDrive: true,
            gui: true,
            cpu: cpuInterval.min,
            ram: ramInterval.min,
            disk: diskInterval.min,
            sharedVolumeMountInfos: [],
          }]
        }));
        form.setFieldsValue({
          templatename: undefined,
        });
      })
      .catch(apolloErrorCatcher);
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
        labelCol={{ span: 2 }}
        wrapperCol={{ span: 22 }}
        form={form}
        onSubmitCapture={onSubmit}
        initialValues={{
          templatename: formTemplate.name,
        }}
      >
        <Form.Item
          {...fullLayout}
          name="templatename"
          className="mt-1"
          required
          validateStatus={valid.name.status as 'success' | 'error'}
          help={valid.name.help}
          validateTrigger="onChange"
          rules={[
            {
              required: true,
              validator: nameValidator,
            },
          ]}
        >
          <Input
            onFocus={() => refetchTemplates({ workspaceNamespace })}
            onChange={e =>
              setFormTemplate(old => {
                return { ...old, name: e.target.value };
              })
            }
            placeholder="Insert template name"
            allowClear
          />
        </Form.Item>

        {/* Environment section */}
        <div className="flex justify-between items-start inline mb-6">
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-lg font-medium">Environments</h3>
            <Button
              type="dashed"
              onClick={addEnvironment}
              icon={<span>+</span>}
              className="text-green-600"
            >
              Add Environment
            </Button>
          </div>

          {formTemplate.environmentList.map((environment, index) => (
            <div key={index} className="my-0">
              <div className="flex justify-between items-center mb-2">
                <h4 className="font-medium">Environment {index + 1}</h4>
                {formTemplate.environmentList.length > 1 && (
                  <Button
                    type="text"
                    danger
                    onClick={() => removeEnvironment(index)}
                    size="small"
                  >
                    Remove
                  </Button>
                )}
              </div>

              {/* Environment Name */}
              <Form.Item
                name={`name_${index}`}
                required
                className="mb-2"
                initialValue={environment.name}
              >
                <Input
                  value={environment.name}
                  onChange={e => updateEnvironment(index, { name: e.target.value })}
                  placeholder="Environment Name"
                  allowClear
                />
              </Form.Item>

              <div className="flex items-end gap-4 mb-3">
                {/* Environment Image */}
                <Form.Item
                  className="my-0"
                  {...fullLayout}
                  style={{ width: '63%' }}
                  name={`image_${index}`}
                  required
                  validateStatus={valid.image.status as 'success' | 'error'}
                  help={valid.image.help}
                  validateTrigger="onChange"
                  rules={[
                    {
                      required: true,
                      validator: imageValidator,
                    },
                  ]}
                >
                  <AutoComplete
                    value={environment.image}
                    options={imagesSearchOptions[index]?.map(x => ({ value: x })) || imagesNoVersion.map(x => ({ value: x }))}
                    onFocus={() => {
                      if (!imagesSearchOptions[index]?.length) {
                        setImagesSearchOptions(old => ({
                          ...old,
                          [index]: imagesNoVersion!
                        }));
                      }
                    }}
                    onChange={value => {
                      setImagesSearchOptions(old => ({
                        ...old,
                        [index]: imagesNoVersion.filter(s => s.includes(value))
                      }));
                      if (value !== environment.image) {
                        const imageFound = images.find(
                          i => getImageNoVer(i.name) === value,
                        );
                        updateEnvironment(index, {
                          image: String(value),
                          registry: imageFound?.registry,
                          vmorcontainer: imageFound?.vmorcontainer[0] ?? EnvironmentType.Container,
                        });
                      }
                    }}
                    placeholder="Select an image"
                  />
                </Form.Item>

                {/* Environment Options */}
                <div className="mt-3">
                  <span>GUI:</span>
                  <Checkbox
                    className="ml-3"
                    checked={environment.gui}
                    onChange={() =>
                      updateEnvironment(index, { gui: !environment.gui })
                    }
                  />
                </div>
                <div className="mr-1 mt-3">
                  <span>Persistent:</span>
                  <Tooltip title="A persistent VM/container disk space won't be destroyed after being turned off.">
                    <Checkbox
                      className="ml-2"
                      checked={environment.persistent}
                      onChange={() =>
                        updateEnvironment(index, {
                          persistent: !environment.persistent,
                          disk: !environment.persistent ? template?.environmentList[index].disk || diskInterval.min : 0
                        })
                      }
                    />
                  </Tooltip>
                </div>
              </div>
              {/* Resource Sliders */}
              <Form.Item labelAlign="left" className="mb-2" label="CPU" name={`cpu_${index}`}>
                <div className="sm:pl-3 pr-1">
                  <Slider
                    styles={{ handle: alternativeHandle }}
                    defaultValue={formTemplate.environmentList[index].cpu}
                    tooltip={{ open: false }}
                    value={environment.cpu}
                    onChange={(value: number) =>
                      updateEnvironment(index, { cpu: value })
                    }
                    min={cpuInterval.min}
                    max={cpuInterval.max}
                    marks={{
                      [cpuInterval.min]: `${cpuInterval.min}`,
                      [environment.cpu]: `${environment.cpu}`,
                      [cpuInterval.max]: `${cpuInterval.max}`,
                    }}
                    included={false}
                    step={1}
                    tipFormatter={(value?: number) => `${value} Core`}
                  />
                </div>
              </Form.Item>
              <Form.Item labelAlign="left" className="mb-2" label="RAM" name={`ram_${index}`}>
                <div className="sm:pl-3 pr-1">
                  <Slider
                    styles={{ handle: alternativeHandle }}
                    defaultValue={environment.ram}
                    tooltip={{ open: false }}
                    value={environment.ram}
                    onChange={(value: number) =>
                      updateEnvironment(index, { ram: value })
                    }
                    min={ramInterval.min}
                    max={ramInterval.max}
                    marks={{
                      [ramInterval.min]: `${ramInterval.min}GB`,
                      [environment.ram]: `${environment.ram}GB`,
                      [ramInterval.max]: `${ramInterval.max}GB`,
                    }}
                    included={false}
                    step={0.25}
                    tipFormatter={(value?: number) => `${value} GB`}
                  />
                </div>
              </Form.Item>

              {environment.persistent && (
                <Form.Item labelAlign="left" className="mb-2" label="DISK" name={`disk_${index}`}  >
                  <div className="sm:pl-3 pr-1">
                    <Slider
                      styles={{ handle: alternativeHandle }}
                      tooltip={{ open: false }}
                      value={environment.disk}
                      defaultValue={environment.disk}
                      onChange={(value: number) =>
                        updateEnvironment(index, { disk: value })
                      }
                      min={diskInterval.min}
                      max={diskInterval.max}
                      marks={{
                        [diskInterval.min]: `${diskInterval.min}GB`,
                        [environment.disk]: `${environment.disk}GB`,
                        [diskInterval.max]: `${diskInterval.max}GB`,
                      }}
                      included={false}
                      step={1}
                      tipFormatter={(value?: number) => `${value} GB`}
                    />
                  </div>
                </Form.Item>
              )}
            </div>
          ))}
        </div>
        <ShVolFormItem workspaceNamespace={workspaceNamespace} />

        <Form.Item {...fullLayout}>
          <div className="flex justify-center">
            {buttonDisabled ? (
              <Tooltip
                title={
                  template
                    ? 'Cannot modify the Template, please change the old parameters and fill all required fields'
                    : 'Cannot create the Template, please fill all required fields'
                }
              >
                <span className="cursor-not-allowed">
                  <Button
                    className="w-24 pointer-events-none"
                    disabled
                    htmlType="submit"
                    type="primary"
                    shape="round"
                    size="middle"
                  >
                    {template ? 'Modify' : 'Create'}
                  </Button>
                </span>
              </Tooltip>
            ) : (
              <Button
                className="w-24"
                htmlType="submit"
                type="primary"
                shape="round"
                size="middle"
                loading={loading}
              >
                {!loading && (template ? 'Modify' : 'Create')}
              </Button>
            )}
          </div>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export type { Template };
export default ModalCreateTemplate;
