import type { FC } from 'react';
import { useState, useEffect, useContext } from 'react';
import {
  Modal,
  Slider,
  Form,
  Input,
  Checkbox,
  Tooltip,
  AutoComplete,
  Select,
  Alert,
} from 'antd';
import { Button } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import type {
  CreateTemplateMutation,
  SharedVolumeMountsListItem,
  ImagesQuery,
} from '../../../generated-types';
import {
  EnvironmentType,
  useWorkspaceTemplatesQuery,
  useImagesQuery,
} from '../../../generated-types';
import type { FetchResult } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import ShVolFormItem, { type ShVolFormItemValue } from './ShVolFormItem';

const alternativeHandle = { border: 'solid 2px #1c7afdd8' };

export type Image = {
  name: string;
  type: Array<ImageType>;
  registry: string;
};

export type ImageList = {
  name: string;
  registryName: string;
  images: Array<{
    name: string;
    versions: Array<string>;
  }>;
};

type ImageType =
  | EnvironmentType.VirtualMachine
  | EnvironmentType.Container
  | EnvironmentType.CloudVm
  | EnvironmentType.Standalone;

type Template = {
  name?: string;
  image?: string;
  registry?: string;
  imageType?: ImageType;
  imageList?: string;
  persistent: boolean;
  mountMyDrive: boolean;
  gui: boolean;
  cpu: number;
  ram: number;
  disk: number;
  sharedVolumeMountInfos?: SharedVolumeMountsListItem[];
};

type Interval = {
  max: number;
  min: number;
};

type Valid = {
  name: { status: string; help?: string };
  image: { status: string; help?: string };
};

// Add this helper function near the top
const getEnvironmentTypeTooltip = (type: ImageType): string => {
  switch (type) {
    case EnvironmentType.Container:
      return 'GUI-based container applications with desktop environment access via web browser. Must follow CrownLabs container guidelines.';
    case EnvironmentType.Standalone:
      return 'Web-based applications exposed over HTTP. Perfect for web services, IDEs, and tools with web interfaces.';
    case EnvironmentType.VirtualMachine:
      return 'Full virtual machines with complete operating system. Supports both GUI and command-line environments.';
    case EnvironmentType.CloudVm:
      return 'Cloud-init compatible virtual machines. SSH access only, no GUI. Suitable for server workloads and CLI applications.';
    default:
      return 'Select the appropriate environment type for your application.';
  }
};

export interface IModalCreateTemplateProps {
  workspaceNamespace: string;
  template?: Template;
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
  isPersonal?: boolean;
}

const getImageNoVer = (image: string) =>
  // split on the last ':' to correctly handle registry:port/repo:tag cases
  image.includes(':') ? image.slice(0, image.lastIndexOf(':')) : image;

const isEmptyOrSpaces = (str: string) => !str || str.match(/^ *$/);

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

  // Process image lists from the query
  const getImageLists = (data: ImagesQuery): ImageList[] => {
    if (!data?.imageList?.images) return [];

    return data.imageList.images
      .filter(img => img?.spec?.registryName && img?.spec?.images)
      .map(img => ({
        name: img!.spec!.registryName,
        registryName: img!.spec!.registryName,
        images: img!
          .spec!.images.filter(i => i?.name && i?.versions)
          .map(i => ({
            name: i!.name,
            versions: i!.versions.filter(v => v !== null) as string[],
          })),
      }));
  };

  // Get images from selected image list
  const getImagesFromList = (imageList: ImageList): Image[] => {
    const images: Image[] = [];

    imageList.images.forEach(img => {
      const versionsInImageName: Image[] = img.versions.map(v => ({
        name: `${img.name}:${v}`,
        type: [],
        registry: imageList.registryName,
      }));

      images.push(...versionsInImageName);
    });

    return images;
  };

  const imageLists = getImageLists(dataImages!);
  const [availableImages, setAvailableImages] = useState<Image[]>([]);

  const [formTemplate, setFormTemplate] = useState<Template>({
    name: template && template.name,
    image: template && template.image,
    registry: template && template.registry,
    imageType: template && template.imageType,
    imageList: template && template.imageList,
    persistent: template?.persistent ?? false,
    mountMyDrive: template?.mountMyDrive ?? true,
    gui: template?.gui ?? true,
    cpu: template ? template.cpu : cpuInterval.min,
    ram: template ? template.ram : ramInterval.min,
    disk: template ? template.disk : diskInterval.min,
    sharedVolumeMountInfos: template ? template.sharedVolumeMountInfos : [],
  });

  // Determine if we're using external images (for non-VM types)
  const isUsingExternalImage =
    formTemplate.imageType &&
    formTemplate.imageType !== EnvironmentType.VirtualMachine;

  // Get available images based on environment type
  const imagesNoVersion = (() => {
    if (formTemplate.imageType === EnvironmentType.VirtualMachine) {
      // For VMs, show images from internal registry
      const baseNames = availableImages.map(x => getImageNoVer(x.name));
      return Array.from(new Set(baseNames)).sort((a, b) => a.localeCompare(b));
    }
    return []; // For other types, no predefined images (external only)
  })();

  const [buttonDisabled, setButtonDisabled] = useState(true);

  const [valid, setValid] = useState<Valid>({
    name: { status: 'success', help: undefined },
    image: { status: 'success', help: undefined },
  });

  const [imagesSearchOptions, setImagesSearchOptions] = useState<string[]>();

  useEffect(() => {
    if (
      formTemplate.name &&
      formTemplate.imageType &&
      valid.name.status === 'success' &&
      // For VMs, check if image is selected from the list
      (formTemplate.imageType === EnvironmentType.VirtualMachine
        ? formTemplate.image && formTemplate.imageList
        : formTemplate.registry) && // For others, check if external image is provided
      (template
        ? template.name !== formTemplate.name ||
          template.image !== formTemplate.image ||
          template.imageType !== formTemplate.imageType ||
          template.imageList !== formTemplate.imageList ||
          template.gui !== formTemplate.gui ||
          template.persistent !== formTemplate.persistent ||
          template.cpu !== formTemplate.cpu ||
          template.ram !== formTemplate.ram ||
          template.disk !== formTemplate.disk ||
          template.registry !== formTemplate.registry ||
          JSON.stringify(template.sharedVolumeMountInfos) !==
            JSON.stringify(formTemplate.sharedVolumeMountInfos)
        : true)
    )
      setButtonDisabled(false);
    else setButtonDisabled(true);
  }, [formTemplate, template, valid.name.status]);

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
    if (formTemplate.imageType === EnvironmentType.VirtualMachine) {
      if (isEmptyOrSpaces(formTemplate.image!)) {
        setValid(old => {
          return {
            ...old,
            image: { status: 'error', help: 'Select an image' },
          };
        });
      } else {
        setValid(old => {
          return {
            ...old,
            image: { status: 'success', help: undefined },
          };
        });
      }
    } else {
      // For external images, validate registry field
      if (isEmptyOrSpaces(formTemplate.registry!)) {
        setValid(old => {
          return {
            ...old,
            image: {
              status: 'error',
              help: 'Enter an external image reference',
            },
          };
        });
      } else {
        setValid(old => {
          return {
            ...old,
            image: { status: 'success', help: undefined },
          };
        });
      }
    }
  };

  const [form] = Form.useForm();

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
    refetch: refetchTemplates,
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

  const onSubmit = () => {
    // prepare sharedVolumeMountInfos for submit (empty for personal templates)
    let sharedVolumeMountInfos: SharedVolumeMountsListItem[] = [];
    if (!isPersonal) {
      const shvolMounts: ShVolFormItemValue[] =
        form.getFieldValue('shvolss') ?? [];
      sharedVolumeMountInfos = (shvolMounts || []).map(obj => ({
        sharedVolume: {
          namespace: String(obj.shvol).split('/')[0],
          name: String(obj.shvol).split('/')[1],
        },
        mountPath: obj.mountpath,
        readOnly: Boolean(obj.readonly),
      }));
    }

    // Determine the final image URL
    let finalImage = '';

    if (formTemplate.imageType === EnvironmentType.VirtualMachine) {
      // For VMs, use the selected image from internal registry
      const selectedImage = availableImages.find(
        i => getImageNoVer(i.name) === formTemplate.image,
      );

      if (selectedImage) {
        finalImage = `registry.internal.crownlabs.polito.it/${selectedImage.name}`;
      } else if (formTemplate.image) {
        finalImage = formTemplate.image.includes('/')
          ? formTemplate.image
          : `registry.internal.crownlabs.polito.it/${formTemplate.image}`;
      }
    } else {
      // For other types, use the external image
      finalImage = formTemplate.registry || '';

      // If it doesn't include a registry, default to internal registry
      if (finalImage && !finalImage.includes('/')) {
        finalImage = `registry.internal.crownlabs.polito.it/${finalImage}`;
      }
    }

    const templateToSubmit = {
      ...formTemplate,
      image: finalImage,
      sharedVolumeMounts: sharedVolumeMountInfos,
    };

    submitHandler(templateToSubmit)
      .then(_result => {
        setShow(false);
        setFormTemplate(old => {
          return {
            ...old,
            name: undefined,
            imageList: undefined,
            image: undefined,
            imageType: undefined,
            registry: undefined,
          };
        });
        setAvailableImages([]);
        form.setFieldsValue({
          templatename: undefined,
          imageList: undefined,
          image: undefined,
          imageType: undefined,
          registry: undefined,
        });
      })
      .catch(error => {
        console.error('ModalCreateTemplate submitHandler error:', error);
        apolloErrorCatcher(error);
      });
  };

  // Environment type options
  const environmentOptions = [
    { value: EnvironmentType.VirtualMachine, label: 'VirtualMachine' },
    { value: EnvironmentType.Container, label: 'Container' },
    { value: EnvironmentType.CloudVm, label: 'CloudVM' },
    { value: EnvironmentType.Standalone, label: 'Standalone' },
  ];

  // Handle environment type selection
  const handleEnvironmentTypeChange = (value: ImageType) => {
    setFormTemplate(old => ({
      ...old,
      imageType: value,
      image: undefined,
      registry: undefined,
      imageList: undefined,
      gui: value === EnvironmentType.CloudVm ? false : true, // CloudVM has no GUI
    }));

    // Reset form fields
    form.setFieldsValue({
      imageType: value,
      image: undefined,
      registry: undefined,
    });

    // For VMs, load the internal registry images
    if (value === EnvironmentType.VirtualMachine) {
      const internalRegistry = imageLists.find(
        list => list.registryName === 'registry.internal.crownlabs.polito.it',
      );

      if (internalRegistry) {
        const images = getImagesFromList(internalRegistry);
        const dedupedImages = images.reduce<Image[]>((acc, img) => {
          const base = getImageNoVer(img.name);
          if (!acc.some(a => getImageNoVer(a.name) === base)) acc.push(img);
          return acc;
        }, []);

        setAvailableImages(dedupedImages);
        setFormTemplate(old => ({
          ...old,
          imageList: 'registry.internal.crownlabs.polito.it',
        }));
      }
    } else {
      // For other types, clear available images
      setAvailableImages([]);
    }

    setImagesSearchOptions(undefined);
  };

  // Handle image selection (for VMs only)
  const handleImageChange = (value: string) => {
    setImagesSearchOptions(imagesNoVersion?.filter(s => s.includes(value)));

    if (value !== formTemplate.image) {
      const imageFound = availableImages.find(
        i => getImageNoVer(i.name) === value,
      );

      setFormTemplate(old => ({
        ...old,
        image: String(value),
        registry: imageFound?.registry,
      }));

      form.setFieldsValue({
        image: value,
      });
    }
  };

  // Initialize available images when editing an existing template
  useEffect(() => {
    if (
      template?.imageType === EnvironmentType.VirtualMachine &&
      imageLists.length
    ) {
      const internalRegistry = imageLists.find(
        list => list.registryName === 'registry.internal.crownlabs.polito.it',
      );

      if (internalRegistry) {
        const imgs = getImagesFromList(internalRegistry);
        const dedupedImgs = imgs.reduce<Image[]>((acc, img) => {
          const base = getImageNoVer(img.name);
          if (!acc.some(a => getImageNoVer(a.name) === base)) acc.push(img);
          return acc;
        }, []);
        setAvailableImages(dedupedImgs);
      }
    }
  }, [template?.imageType, imageLists]);

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
          imageType: formTemplate.imageType,
          image: formTemplate.image,
          registry: formTemplate.registry,
          cpu: formTemplate.cpu,
          ram: formTemplate.ram,
          disk: formTemplate.disk,
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
            onFocus={() => {
              refetchTemplates({ workspaceNamespace });
            }}
            onChange={e =>
              setFormTemplate(old => {
                return { ...old, name: e.target.value };
              })
            }
            placeholder="Insert template name"
            allowClear
          />
        </Form.Item>

        {/* Environment Type Selection - Now comes first */}
        <Form.Item
          {...fullLayout}
          label="Environment Type"
          name="imageType"
          className="mb-4"
          required
          rules={[
            {
              required: true,
              message: 'Please select an environment type',
            },
          ]}
        >
          <Select
            value={formTemplate.imageType}
            onChange={handleEnvironmentTypeChange}
            placeholder="Select environment type"
          >
            {environmentOptions.map(option => (
              <Select.Option key={option.value} value={option.value}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                  }}
                >
                  <span>{option.label}</span>
                  <Tooltip title={getEnvironmentTypeTooltip(option.value)}>
                    <InfoCircleOutlined
                      style={{ color: '#1890ff', marginLeft: 8 }}
                    />
                  </Tooltip>
                </div>
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        {/* VM Image Selection */}
        {formTemplate.imageType === EnvironmentType.VirtualMachine && (
          <Form.Item
            className="mb-4"
            {...fullLayout}
            name="image"
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
              options={(imagesSearchOptions ?? imagesNoVersion).map(x => ({
                value: x,
              }))}
              onFocus={() => {
                if (!imagesSearchOptions?.length)
                  setImagesSearchOptions(imagesNoVersion);
              }}
              onChange={handleImageChange}
              placeholder="Select a virtual machine image"
            />
          </Form.Item>
        )}

        {/* External Image Input for Container, CloudVM, Standalone */}
        {isUsingExternalImage && (
          <>
            {/* Information section for external image requirements */}
            <Alert
              message={`${formTemplate.imageType} Image Requirements`}
              description={
                <div>
                  {formTemplate.imageType === EnvironmentType.Container && (
                    <p>
                      Must be compliant with{' '}
                      <a
                        href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/containers"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        CrownLabs container guidelines
                      </a>
                      . GUI-based container applications with desktop
                      environment access via web browser.
                    </p>
                  )}
                  {formTemplate.imageType === EnvironmentType.Standalone && (
                    <p>
                      Must be compliant with{' '}
                      <a
                        href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/standalone"
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        CrownLabs standalone guidelines
                      </a>
                      . Web-based applications exposed over HTTP, perfect for
                      web services, IDEs, and tools with web interfaces.
                    </p>
                  )}
                  {formTemplate.imageType === EnvironmentType.CloudVm && (
                    <p>
                      Can be any cloud-init compatible image, but will only be
                      accessible via SSH{' '}
                      <Tooltip title="CloudVM images must support cloud-init and will not have GUI access - SSH access only">
                        <InfoCircleOutlined style={{ color: '#1890ff' }} />
                      </Tooltip>
                      . Suitable for server workloads and CLI applications.
                    </p>
                  )}
                </div>
              }
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />

            {/* External Image Input */}
            <Form.Item
              {...fullLayout}
              label="External Image"
              name="registry"
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
              extra="Examples: ubuntu:22.04, docker.io/library/nginx:latest, registry.internal.crownlabs.polito.it/netgroup/ubuntu-server-base:20200922"
            >
              <Input
                value={formTemplate.registry}
                onChange={e =>
                  setFormTemplate(old => ({
                    ...old,
                    registry: e.target.value,
                  }))
                }
                placeholder="Enter image name (e.g., ubuntu:22.04)"
                suffix={
                  <Tooltip title="Image format: [registry/]repository[:tag]">
                    <InfoCircleOutlined style={{ color: 'rgba(0,0,0,.45)' }} />
                  </Tooltip>
                }
              />
            </Form.Item>
          </>
        )}

        <div className="flex justify-between items-start inline mb-6">
          <Form.Item className="mb-4">
            <span>GUI:</span>
            <Checkbox
              className="ml-3"
              checked={formTemplate.gui}
              disabled={formTemplate.imageType === EnvironmentType.CloudVm} // Disable GUI for CloudVM
              onChange={() =>
                setFormTemplate(old => {
                  return { ...old, gui: !old.gui };
                })
              }
            />
            {formTemplate.imageType === EnvironmentType.CloudVm && (
              <div
                style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}
              >
                CloudVM instances do not support GUI access
              </div>
            )}
          </Form.Item>

          <Form.Item className="mb-4">
            <span>Persistent: </span>
            <Tooltip title="A persistent VM/container disk space won't be destroyed after being turned off.">
              <Checkbox
                className="ml-3"
                checked={formTemplate.persistent}
                onChange={() =>
                  setFormTemplate(old => {
                    return {
                      ...old,
                      persistent: !old.persistent,
                      disk: !old.persistent
                        ? template?.disk || diskInterval.min
                        : 0,
                    };
                  })
                }
              />
            </Tooltip>
          </Form.Item>
        </div>

        <Form.Item labelAlign="left" className="mt-10" label="CPU" name="cpu">
          <div className="sm:pl-3 pr-1">
            <Slider
              styles={{ handle: alternativeHandle }}
              defaultValue={formTemplate.cpu}
              tooltip={{ open: false }}
              value={formTemplate.cpu}
              onChange={(value: number) =>
                setFormTemplate(old => {
                  return { ...old, cpu: value };
                })
              }
              min={cpuInterval.min}
              max={cpuInterval.max}
              marks={{
                [cpuInterval.min]: `${cpuInterval.min}`,
                [formTemplate.cpu]: `${formTemplate.cpu}`,
                [cpuInterval.max]: `${cpuInterval.max}`,
              }}
              included={false}
              step={1}
              tipFormatter={(value?: number) => `${value} Core`}
            />
          </div>
        </Form.Item>
        <Form.Item labelAlign="left" label="RAM" name="ram">
          <div className="sm:pl-3 pr-1">
            <Slider
              styles={{ handle: alternativeHandle }}
              defaultValue={formTemplate.ram}
              tooltip={{ open: false }}
              value={formTemplate.ram}
              onChange={(value: number) =>
                setFormTemplate(old => {
                  return { ...old, ram: value };
                })
              }
              min={ramInterval.min}
              max={ramInterval.max}
              marks={{
                [ramInterval.min]: `${ramInterval.min}GB`,
                [formTemplate.ram]: `${formTemplate.ram}GB`,
                [ramInterval.max]: `${ramInterval.max}GB`,
              }}
              included={false}
              step={0.25}
              tipFormatter={(value?: number) => `${value} GB`}
            />
          </div>
        </Form.Item>
        <Form.Item
          labelAlign="left"
          label="DISK"
          name="disk"
          className={formTemplate.persistent ? '' : 'hidden'}
        >
          <div className="sm:pl-3 pr-1 ">
            <Slider
              styles={{ handle: alternativeHandle }}
              tooltip={{ open: false }}
              value={formTemplate.disk}
              defaultValue={formTemplate.disk}
              onChange={(value: number) =>
                setFormTemplate(old => {
                  return { ...old, disk: value };
                })
              }
              min={diskInterval.min}
              max={diskInterval.max}
              marks={{
                [diskInterval.min]: `${diskInterval.min}GB`,
                [formTemplate.disk]: `${formTemplate.disk}GB`,
                [diskInterval.max]: `${diskInterval.max}GB`,
              }}
              included={false}
              step={1}
              tipFormatter={(value?: number) => `${value} GB`}
            />
          </div>
        </Form.Item>

        {!isPersonal && (
          <ShVolFormItem workspaceNamespace={workspaceNamespace} />
        )}

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
