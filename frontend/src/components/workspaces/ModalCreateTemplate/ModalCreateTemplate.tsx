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
      // Don't set a default type - let users choose
      const versionsInImageName: Image[] = img.versions.map(v => ({
        name: `${img.name}:${v}`,
        type: [], // No default type - user must choose
        registry: imageList.registryName,
      }));

      images.push(...versionsInImageName);
    });

    return images;
  };

  const imageLists = getImageLists(dataImages!);
  const [selectedImageList, setSelectedImageList] = useState<ImageList | null>(
    null,
  );
  const [availableImages, setAvailableImages] = useState<Image[]>([]);

  // Determine if the selected image list contains container images
  const isContainerImageList =
    selectedImageList?.registryName === 'crownlabs-container-envs' ||
    selectedImageList?.registryName === 'registry.internal.crownlabs.polito.it';

  // const isStandaloneImageList =
  //   selectedImageList?.registryName === 'crownlabs-standalone';

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

  // Move this before the useEffect that uses it
  const isExternalImage = formTemplate.image === '**-- External image --**';

  // Add "External image" to the options only if personal workspace AND container image list is selected
  // Deduplicate base names (remove duplicates caused by multiple versions) and sort for stable display
  const imagesNoVersion = (() => {
    const external =
      isPersonal && isContainerImageList ? ['**-- External image --**'] : [];
    const baseNames = availableImages.map(x => getImageNoVer(x.name));
    const uniqueSorted = Array.from(new Set(baseNames)).sort((a, b) =>
      a.localeCompare(b),
    );
    return [...external, ...uniqueSorted];
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
      formTemplate.image &&
      formTemplate.imageType &&
      formTemplate.imageList &&
      valid.name.status === 'success' &&
      // For external images, also check that registry is provided
      (!isExternalImage || (isExternalImage && formTemplate.registry)) &&
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
  }, [formTemplate, template, valid.name.status, isExternalImage]);

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
    if (isEmptyOrSpaces(formTemplate.image!)) {
      setValid(old => {
        return {
          ...old,
          image: { status: 'error', help: 'Insert an image' },
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

    if (formTemplate.image === '**-- External image --**') {
      // For external images, use the registry field as the complete image URL
      finalImage = formTemplate.registry || '';

      // If it doesn't include a registry, default to internal registry
      if (finalImage && !finalImage.includes('/')) {
        finalImage = `registry.internal.crownlabs.polito.it/${finalImage}`;
      }
    } else {
      // For selected images from image lists
      const selectedImage = availableImages.find(
        i => getImageNoVer(i.name) === formTemplate.image,
      );

      if (selectedImage) {
        // Images from the internal registry already have the full name
        // Just need to add the registry prefix
        finalImage = `registry.internal.crownlabs.polito.it/${selectedImage.name}`;
      } else if (formTemplate.image) {
        // Fallback for any other case
        finalImage = formTemplate.image.includes('/')
          ? formTemplate.image
          : `registry.internal.crownlabs.polito.it/${formTemplate.image}`;
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
          };
        });
        setSelectedImageList(null);
        setAvailableImages([]);
        form.setFieldsValue({
          templatename: undefined,
          imageList: undefined,
          image: undefined,
        });
      })
      .catch(error => {
        console.error('ModalCreateTemplate submitHandler error:', error);
        apolloErrorCatcher(error);
      });
  };

  // Helper function to map registry names to URLs
  // const getRegistryUrl = (registryName: string): string => {
  //   // Since we only have one registry, just return it as-is
  //   if (registryName === 'registry.internal.crownlabs.polito.it') {
  //     return registryName;
  //   }
  //   // For any other registry, return as-is
  //   return registryName;
  // };

  // Environment type options for external images
  const environmentOptions = [
    { value: EnvironmentType.VirtualMachine, label: 'VirtualMachine' },
    { value: EnvironmentType.Container, label: 'Container' },
    { value: EnvironmentType.CloudVm, label: 'CloudVM' },
    { value: EnvironmentType.Standalone, label: 'Standalone' },
  ];

  // Handle image list selection
  const handleImageListChange = (imageListName: string) => {
    const selectedList = imageLists.find(list => list.name === imageListName);
    if (selectedList) {
      setSelectedImageList(selectedList);
      const images = getImagesFromList(selectedList);

      // dedupe images by base name (remove entries that differ only by version)
      const dedupedImages = images.reduce<Image[]>((acc, img) => {
        const base = getImageNoVer(img.name);
        if (!acc.some(a => getImageNoVer(a.name) === base)) acc.push(img);
        return acc;
      }, []);

      setAvailableImages(dedupedImages);
      setImagesSearchOptions(undefined);

      // Reset form when changing image list
      setFormTemplate(old => ({
        ...old,
        imageList: imageListName,
        image: undefined,
        registry: undefined,
        imageType: undefined, // Reset imageType so user must choose
        gui: true, // Default to GUI enabled
      }));

      form.setFieldsValue({
        imageList: imageListName,
        image: undefined,
        imageType: undefined,
      });
    }
  };

  // Handle image selection
  const handleImageChange = (value: string) => {
    // Update search options as user types
    setImagesSearchOptions(imagesNoVersion?.filter(s => s.includes(value)));

    if (value !== formTemplate.image) {
      if (value === '**-- External image --**') {
        setFormTemplate(old => ({
          ...old,
          image: '**-- External image --**',
          registry: '', // reset registry
          imageType: undefined, // Let user choose type
          persistent: false,
          gui: true,
        }));
        form.setFieldsValue({
          image: value,
          imageType: undefined,
        });
      } else {
        const imageFound = availableImages.find(
          i => getImageNoVer(i.name) === value,
        );
        setFormTemplate(old => ({
          ...old,
          image: String(value),
          registry: imageFound?.registry,
          imageType: undefined, // Let user choose type for all images
          persistent: false,
          gui: true,
        }));
        form.setFieldsValue({
          image: value,
          imageType: undefined,
        });
      }
    }
  };

  // Initialize selected list & available images when editing an existing template
  useEffect(() => {
    if (template?.imageList && imageLists.length) {
      const selected = imageLists.find(l => l.name === template.imageList);
      if (selected) {
        setSelectedImageList(selected);
        const imgs = getImagesFromList(selected);
        const dedupedImgs = imgs.reduce<Image[]>((acc, img) => {
          const base = getImageNoVer(img.name);
          if (!acc.some(a => getImageNoVer(a.name) === base)) acc.push(img);
          return acc;
        }, []);
        setAvailableImages(dedupedImgs);
      }
    }
  }, [template?.imageList, imageLists]);

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
          imageList: formTemplate.imageList,
          image: formTemplate.image,
          imageType: formTemplate.imageType,
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

        <Form.Item
          {...fullLayout}
          name="imageList"
          className="mb-4"
          required
          rules={[{ required: true, message: 'Please select an image list' }]}
        >
          <Select
            placeholder="Select an image list"
            value={formTemplate.imageList}
            onChange={handleImageListChange}
            options={imageLists.map(list => ({
              value: list.name,
              label: list.name,
            }))}
          />
        </Form.Item>

        <div className="flex justify-between items-start inline mb-6">
          <Form.Item
            className="mb-4"
            {...fullLayout}
            style={{ width: '100%' }}
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
              disabled={!selectedImageList}
              onFocus={() => {
                if (!imagesSearchOptions?.length)
                  setImagesSearchOptions(imagesNoVersion);
              }}
              onChange={handleImageChange}
              placeholder={
                !selectedImageList
                  ? 'Select an image list first'
                  : 'Select an image'
              }
            />
          </Form.Item>

          {/* Environment Type Selection - Always show when image is selected */}
          {formTemplate.image &&
            formTemplate.image !== '**-- External image --**' && (
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
                  onChange={(value: ImageType) =>
                    setFormTemplate(old => ({
                      ...old,
                      imageType: value,
                    }))
                  }
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
                        <Tooltip
                          title={getEnvironmentTypeTooltip(option.value)}
                        >
                          <InfoCircleOutlined
                            style={{ color: '#1890ff', marginLeft: 8 }}
                          />
                        </Tooltip>
                      </div>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            )}

          {isExternalImage && (
            <>
              {/* Information section for external image requirements */}
              <Alert
                message="External Image Requirements"
                description={
                  <div>
                    <p>
                      When using external images, please ensure your
                      container/image complies with the appropriate CrownLabs
                      requirements:
                    </p>
                    <ul style={{ marginBottom: 0, paddingLeft: '20px' }}>
                      <li>
                        <strong>Container:</strong> Must be compliant with{' '}
                        <a
                          href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/containers"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          CrownLabs container guidelines
                        </a>
                      </li>
                      <li>
                        <strong>Standalone:</strong> Must be compliant with{' '}
                        <a
                          href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/standalone"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          CrownLabs standalone guidelines
                        </a>
                      </li>
                      <li>
                        <strong>Virtual Machine:</strong> Must be compliant with{' '}
                        <a
                          href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/virtual-machines"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          CrownLabs VM guidelines
                        </a>
                      </li>
                      <li>
                        <strong>CloudVM:</strong> Can be any cloud-init
                        compatible image, but will only be accessible via SSH{' '}
                        <Tooltip title="CloudVM images must support cloud-init and will not have GUI access - SSH access only">
                          <InfoCircleOutlined style={{ color: '#1890ff' }} />
                        </Tooltip>
                      </li>
                    </ul>
                  </div>
                }
                type="info"
                showIcon
                style={{ marginBottom: 16 }}
              />

              {/* Environment Type Selection for External Images */}
              <Form.Item
                label="Environment Type"
                name="imageType"
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
                  onChange={(value: ImageType) =>
                    setFormTemplate(old => ({
                      ...old,
                      imageType: value,
                    }))
                  }
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
                        <Tooltip
                          title={getEnvironmentTypeTooltip(option.value)}
                        >
                          <InfoCircleOutlined
                            style={{ color: '#1890ff', marginLeft: 8 }}
                          />
                        </Tooltip>
                      </div>
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>

              {/* External Image Input */}
              <Form.Item
                label="Container Image"
                name="registry"
                required
                rules={[
                  {
                    required: true,
                    message: 'Please enter the container image',
                  },
                  {
                    pattern:
                      /^([a-z0-9]+([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]+([-a-z0-9]*[a-z0-9])?)*(:[0-9]+)?\/)?(([a-z0-9]+([-._]?[a-z0-9]+)*\/)*[a-z0-9]+([-._]?[a-z0-9]+)*)(:[a-zA-Z0-9]+([-._]?[a-zA-Z0-9]+)*)?$|^[a-z0-9]+([-._]?[a-z0-9]+)*\/[a-z0-9]+([-._]?[a-z0-9]+)*(:[a-zA-Z0-9]+([-._]?[a-z0-9]+)*)?$/,
                    message: 'Please enter a valid container image name',
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
                      <InfoCircleOutlined
                        style={{ color: 'rgba(0,0,0,.45)' }}
                      />
                    </Tooltip>
                  }
                />
              </Form.Item>
            </>
          )}

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
