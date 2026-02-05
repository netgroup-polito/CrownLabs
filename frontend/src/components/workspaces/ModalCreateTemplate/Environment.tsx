import { InfoCircleOutlined } from '@ant-design/icons';
import {
  Alert,
  AutoComplete,
  Checkbox,
  Form,
  Input,
  Select,
  Slider,
  Flex,
  Tooltip,
  InputNumber,
  Space
} from 'antd';
import { useEffect, useState, type FC } from 'react';
import { EnvironmentType } from '../../../generated-types';
import { SharedVolumeList } from './SharedVolumeList';
import type { SharedVolume } from '../../../utils';
import type { ChildFormItem, Resources, TemplateFormEnv, Image } from './types';
import { EnvironmentDisk } from './EnvironmentDisk';
import { formItemLayout, getImageNameNoVer } from './utils';

// Environment type options
const environmentTypeOptions = [
  { value: EnvironmentType.VirtualMachine, label: 'Virtual Machine' },
  { value: EnvironmentType.CloudVm, label: 'Cloud VM' },
  { value: EnvironmentType.Standalone, label: 'Standalone' },
  { value: EnvironmentType.Container, label: 'Container' },
];

const getImageNames = (images: Image[]) => {
  const baseNames = images.map(img => getImageNameNoVer(img.name));
  return Array.from(new Set(baseNames)).sort((a, b) => a.localeCompare(b));
};

type EnvironmentProps = {
  availableImages: Image[];
  resources: Resources;
  sharedVolumes: SharedVolume[];
  isPersonal: boolean;
} & ChildFormItem;

export const Environment: FC<EnvironmentProps> = ({
  parentFormName: name,
  restField,
  availableImages,
  resources,
  sharedVolumes,
  isPersonal,
}) => {
  const form = Form.useFormInstance();

  const environments = Form.useWatch<TemplateFormEnv[] | undefined>(
    'environments',
  );

  const propInputField = {
    labelCol: { span: 24 },
    wrapperCol: { span: 24 },
    className: "mb-0",
  }

  // Custom validator for unique environment names
  const validateUniqueName = (currIndex: number) => {
    return async (_: unknown, name: string) => {
      if (!environments || !name) return;

      const trimmedName = name.trim().toLowerCase();
      const duplicateIndex = environments.findIndex(
        (env, idx) =>
          idx !== currIndex && env.name.toLowerCase() === trimmedName,
      );

      if (duplicateIndex !== -1) {
        throw new Error(`Name "${name}" is already used`);
      }
    };
  };

  // Function to trigger validation of all name fields when any name changes
  const handleNameChange = (changedIndex: number) => {
    if (!environments) return;

    // Validate all other name fields to update their validation status
    environments.forEach((_: TemplateFormEnv, idx: number) => {
      if (idx !== changedIndex) {
        form.validateFields([['environments', idx, 'name']]).catch(() => {
          // Ignore validation errors, just trigger the validation
        });
      }
    });
  };

  const validateVMImageName = async (_: unknown, image: string) => {
    if (!environments || !image) return;

    // Check if the image is in the list of available images
    const found = getImageNames(availableImages).find(
      tmpImage => tmpImage === image,
    );
    if (found) return;

    throw new Error(`Image "${image}" is not found from registry`);
  };

  const [imagesSearchOptions, setImagesSearchOptions] = useState<string[]>([]);

  useEffect(() => {
    setImagesSearchOptions(getImageNames(availableImages));
  }, [availableImages]);

  const isVM = (currIndex: number) => {
    if (!environments) return false;
    if (!environments[currIndex]) return false;

    return (
      environments[currIndex].environmentType === EnvironmentType.VirtualMachine
    );
  };

  const isPersistent = (currIndex: number) => {
    if (!environments) return false;
    if (!environments[currIndex]) return false;

    return environments[currIndex].persistent;
  };

  const getImageAlert = (currIndex: number) => {
    if (!environments) return <></>;
    if (!environments[currIndex]) return <></>;
    if (!environments[currIndex].environmentType) return <></>;

    switch (environments[currIndex].environmentType) {
      case EnvironmentType.CloudVm:
        return <CloudVmAlert />;
      case EnvironmentType.Container:
        return <ContainerAlert />;
      case EnvironmentType.Standalone:
        return <StandaloneAlert />;
    }

    return <></>;
  };

  const getGUIDescription = (currIndex: number) => {
    if (!environments) return '';
    if (!environments[currIndex]) return '';

    switch (environments[currIndex].environmentType) {
      case EnvironmentType.Container:
      case EnvironmentType.Standalone:
        return 'Standalone and Container environments only work with GUI and not SSH access';
      case EnvironmentType.CloudVm:
        return 'CloudVM instances do not support GUI access';
    }

    return '';
  };

  const getEnvironmentType = (currIndex: number) => {
    if (!environments) return '';
    if (!environments[currIndex]) return '';

    return environments[currIndex].environmentType;
  };

  const handleEnvTypeChange = (envIndex: number, envType: EnvironmentType) => {
    if (!environments) return;
    if (!environments[envIndex]) return;

    form.setFieldsValue({
      environments: environments.map((env, idx) => {
        if (idx === envIndex) {
          let gui = env.gui;
          let rewriteUrl = env.rewriteUrl;
          let persistent = env.persistent;
          let disk = env.disk;

          switch (envType) {
            case EnvironmentType.Container:
              gui = true;
              rewriteUrl = false;
              break;

            case EnvironmentType.Standalone:
              gui = true;
              rewriteUrl = true;
              break;

            case EnvironmentType.CloudVm:
              gui = false;
              rewriteUrl = false;
              persistent = true;
              disk = Math.max(disk, resources.disk.min);
              break;

            case EnvironmentType.VirtualMachine:
              rewriteUrl = false;
              break;
          }

          return {
            ...env,
            environmentType: envType,
            gui: gui,
            rewriteUrl: rewriteUrl,
            persistent: persistent,
            disk: disk,
          };
        }
        return env;
      }),
    });
  };

  const handleResourceChange = (
    currIndex: number,
    field: 'cpu' | 'ram' | 'reservedCpu',
    value: number | null,
  ) => {
    if (value === null) return;
    if (!environments) return;
    if (!environments[currIndex]) return;

    form.setFieldsValue({
      environments: environments.map((env, idx) => {
        if (idx === currIndex) {
          return {
            ...env,
            [field]: value,
          };
        }
        return env;
      }),
    });
  };

  const getExternalImageExample = (currIndex: number): string | undefined => {
    if (!environments) return undefined;
    if (!environments[currIndex]) return undefined;

    switch (environments[currIndex].environmentType) {
      case EnvironmentType.Container:
        return 'Examples: ubuntu:22.04, docker.io/library/nginx:latest';
      case EnvironmentType.Standalone:
        return 'Example: crownlabs/vscode-rust:v0.2.0';
      case EnvironmentType.CloudVm:
        return 'Example: https://cloud-images.ubuntu.com/minimal/releases/noble/release/ubuntu-24.04-minimal-cloudimg-amd64.img';
      default:
        return undefined;
    }
  };

  const getExternalImagePlaceholder = (currIndex: number): string => {
    if (!environments) return 'Enter image name';
    if (!environments[currIndex]) return 'Enter image name';

    switch (environments[currIndex].environmentType) {
      case EnvironmentType.Container:
        return 'Enter image name (e.g., ubuntu:22.04)';
      case EnvironmentType.Standalone:
        return 'Enter image name (e.g., crownlabs/vscode-rust:v0.2.0)';
      case EnvironmentType.CloudVm:
        return 'Enter image URL (e.g., https://cloud-images.ubuntu.com/...)';
      default:
        return 'Enter image name';
    }
  };

    const handlePersistentChange = (checked: boolean) => {
    if (!environments) return;
    if (!environments[name]) return;

    form.setFieldsValue({
      environments: environments.map((env, idx) => {
        if (idx === name) {
          return {
            ...env,
            persistent: checked,
            disk: checked ? resources.disk.min : 0,
          };
        }
        return env;
      }),
    });
  };

  return (
    <>
      {/* Environment Name */}
      <Form.Item
        {...restField}
        name={[name, 'name']}
        label="Name"
        validateTrigger={['onChange', 'onBlur']}
        rules={[
          { required: true, message: 'Environment name is required' },
          { 
            pattern: /^[a-z\d][a-z\d-]{2,10}[a-z\d]$/,
            message: 'Name must be 4-12 characters: lowercase letters, digits, hyphens (no hyphens at start/end)'
          },
          { validator: validateUniqueName(name) },
        ]}
        validateDebounce={500}
        {...formItemLayout}
      >
        <Input
          placeholder="Environment Name"
          allowClear
          onChange={() => handleNameChange(name)}
        />
      </Form.Item>

      {/* Environment Type Selection */}
      <Form.Item
        label="Type"
        name={[name, 'environmentType']}
        required
        {...formItemLayout}
      >
        <Select
          placeholder="Select environment type"
          getPopupContainer={trigger => trigger.parentElement || document.body}
          onChange={value =>
            handleEnvTypeChange(name, value as EnvironmentType)
          }
        >
          {environmentTypeOptions.map(option => (
            <Select.Option key={option.value} value={option.value}>
              {option.label}
            </Select.Option>
          ))}
        </Select>
      </Form.Item>

      {/* VM Image Selection - Remove {...fullLayout} */}
      {isVM(name) ? (
        <Form.Item
          {...restField}
          label="Image"
          name={[name, 'image']}
          required
          validateTrigger="onChange"
          rules={[
            {
              required: true,
              message: 'Select a virtual machine image',
            },
            {
              validator: validateVMImageName,
            },
          ]}
          validateDebounce={500}
          {...formItemLayout}
        >
          <AutoComplete
            options={imagesSearchOptions.map(imgName => ({
              value: imgName,
            }))}
            onFocus={() => {
              if (imagesSearchOptions.length === 0)
                setImagesSearchOptions(getImageNames(availableImages));
            }}
            onChange={(value: string) => {
              setImagesSearchOptions(
                getImageNames(availableImages).filter(s => s.includes(value)),
              );
            }}
            placeholder="Select a virtual machine image"
            getPopupContainer={trigger =>
              trigger.parentElement || document.body
            }
          />
        </Form.Item>
      ) : (
        <>
          {/* External Image Input for Container, CloudVM, Standalone */}
          <Alert
            message={`${getEnvironmentType(name)} Image Requirements`}
            description={getImageAlert(name)}
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          {/* External Image Input */}
          <Form.Item
            {...restField}
            label="External Image"
            name={[name, 'image']}
            required
            validateTrigger="onChange"
            labelCol={{ span: 6 }}
            rules={[
              {
                required: true,
                message: 'Enter an external image',
              },
            ]}

            extra={getExternalImageExample(name)}
          >
            <Input
              placeholder={getExternalImagePlaceholder(name)}
              suffix={
                <Tooltip title="Image format: [registry/]repository[:tag] or URL for CloudVM">
                  <InfoCircleOutlined style={{ color: 'rgba(0,0,0,.45)' }} />
                </Tooltip>
              }
            />
          </Form.Item>
        </>
      )}
      
    <Flex vertical={!isVM(name)} {...isVM(name)&& {justify: "space-evenly"}} className='ml-6 gap-2'>
      {/* GUI Toggle */}
      <Form.Item label="GUI" className="mb-0">
        <div className="flex items-center gap-2">
          <Form.Item
            {...restField}
            name={[name, 'gui']}
            valuePropName="checked"
            className="mb-0"
          >
            <Checkbox
              disabled={!isVM(name)}
            />
          </Form.Item>

          <div className="ant-form-item-extra text-xs">
            {getGUIDescription(name)}
          </div>
        </div>
      </Form.Item>

      {/* Persistent Checkbox */}
      <Tooltip title="A persistent VM/container disk space won't be destroyed after being turned off.">
        <Form.Item label="Persistent" className="mb-0">
          <div className="flex items-center ">
            <Form.Item
              {...restField}
              name={[name, 'persistent']}
              valuePropName="checked"
              className="mb-0"
            >
              <Checkbox
                disabled={getEnvironmentType(name) === EnvironmentType.CloudVm}
                onChange={value => handlePersistentChange(value.target.checked)}
              />
            </Form.Item>
          </div>
        </Form.Item>
      </Tooltip>

      {/* Rewrite URL toggle - only for Standalone */}
      {getEnvironmentType(name) === EnvironmentType.Standalone && (
        <Form.Item label="Rewrite URL" className="mb-0">
          <div className="flex items-center gap-2">
            <Form.Item
              {...restField}
              name={[name, 'rewriteUrl']}
              valuePropName="checked"
              className="mb-0"
            >
              <Checkbox disabled />
            </Form.Item>

            <div className="ant-form-item-extra text-xs">
              Rewrite incoming URLs to the application URL (required for Standalone).
            </div>
          </div>
        </Form.Item>
      )}
    </Flex>
      
      <Space direction="horizontal" className='ml-2 mb-4'  >
      {/* CPU Input */}
      <Form.Item 
        {...restField}
        label="CPU"
        name={[name, 'cpu']}
        {...propInputField}
      > 
        <InputNumber
          step={0.5}
          style={{width:"120px", textAlignLast: "center"}}
          min={resources.cpu.min}
          max={resources.cpu.max}
          defaultValue={form.getFieldValue(['environments', name, 'cpu']) || 1.0}
          onChange={value => handleResourceChange(name, 'cpu', value)}
          addonAfter="vCPU"
        />
      </Form.Item>


      {/* RAM Input */}
      <Form.Item
        {...restField}
        name={[name, 'ram']}
        label="RAM"
        {...propInputField}
      >
        <InputNumber
          style={{width:"120px", textAlignLast: "center"}}
          min={resources.ram.min}
          max={resources.ram.max}
          defaultValue={form.getFieldValue(['environments', name, 'ram']) || 1.0}
          step={0.25}
          onChange={value => handleResourceChange(name, 'ram', value)}
          addonAfter="GB"
        />
      </Form.Item>

      {/* Disk */}
        {/* TODO: The minimum disk size should be dynamically set based on the VM image metadata. Right now, if instance is a VM, min 10; otherwise minimum 1 */}
        <Form.Item {...restField} name={[name,'disk']} label="Disk" {...propInputField} >
            <InputNumber
              step={1}
              style={{ width: "120px", textAlignLast: "center" }}
              
              addonAfter="GB"
              disabled={!isPersistent(name)}
              max={resources.disk.max}
              min={getEnvironmentType(name) === EnvironmentType.VirtualMachine ? isPersistent(name) ? resources.disk.min : 0 : 0}
            />
          </Form.Item>

        {/* Reserved CPU Percentage */}
        <Form.Item {...restField} name={[name,'reservedCpu']} label="Reserved CPU" {...propInputField} >
            <InputNumber
              onChange={value => handleResourceChange(name, 'reservedCpu', value)}
              step={1}
              style={{ width: "120px", textAlignLast: "center" }}
              defaultValue={50}
              addonAfter="%"
              max={100}
              min={1}
            />
          </Form.Item>
      </Space>

      {/* Persistance/Disk */}
      {/* <EnvironmentDisk
        parentFormName={name}
        restField={restField}
        diskResources={resources.disk}
        isCloudVm={getEnvironmentType(name) === EnvironmentType.CloudVm}
      /> */}

      {!isPersonal && (
        <SharedVolumeList parentFormName={name} sharedVolumes={sharedVolumes}  />
      )}
    </>
  );
};

const ContainerAlert = () => {
  return (
    <p>
      Must be compliant with{' '}
      <a
        href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/containers"
        target="_blank"
        rel="noopener noreferrer"
      >
        CrownLabs container guidelines
      </a>
      . GUI-based container applications with desktop environment access via web
      browser.
    </p>
  );
};

const StandaloneAlert = () => {
  return (
    <p>
      Must be compliant with{' '}
      <a
        href="https://github.com/netgroup-polito/CrownLabs/tree/master/provisioning/standalone"
        target="_blank"
        rel="noopener noreferrer"
      >
        CrownLabs standalone guidelines
      </a>
      . Web-based applications exposed over HTTP, perfect for web services,
      IDEs, and tools with web interfaces.
    </p>
  );
};

const CloudVmAlert = () => {
  return (
    <p>
      Can be any cloud-init compatible image, but will only be accessible via
      SSH. It requires an appropriate disk and it must be persistent. Suitable
      for server workloads and CLI applications.
    </p>
  );
};
