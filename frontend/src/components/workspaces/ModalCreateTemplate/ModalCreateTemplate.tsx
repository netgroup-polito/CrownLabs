import { useState, useEffect, FC, useContext } from 'react';
import {
  Modal,
  Slider,
  Form,
  Input,
  Checkbox,
  Tooltip,
  AutoComplete,
} from 'antd';
import Button from 'antd-button-color';
import {
  CreateTemplateMutation,
  useWorkspaceTemplatesQuery,
} from '../../../generated-types';
import { FetchResult } from 'apollo-link';
import { ErrorContext } from '../../../errorHandling/ErrorContext';

const alternativeHandle = { border: 'solid 2px #1c7afdd8' };

export type Image = {
  name: string;
  vmorcontainer: Array<Vmorcontainer>;
  registry: string;
};

type Vmorcontainer = 'Container' | 'VM';

type Template = {
  name?: string;
  image?: string;
  registry?: string;
  vmorcontainer?: Vmorcontainer;
  persistent: boolean;
  mountMyDrive: boolean;
  gui: boolean;
  cpu: number;
  ram: number;
  disk: number;
};

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
    t: Template
  ) => Promise<
    FetchResult<
      CreateTemplateMutation,
      Record<string, any>,
      Record<string, any>
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
    name: template && template.name,
    image: template && template.image,
    registry: template && template.registry,
    vmorcontainer: template && template.vmorcontainer,
    persistent: template?.persistent ?? false,
    mountMyDrive: template?.mountMyDrive ?? true,
    gui: template?.gui ?? true,
    cpu: template ? template.cpu : cpuInterval.min,
    ram: template ? template.ram : ramInterval.min,
    disk: template ? template.disk : diskInterval.min,
  });

  const [valid, setValid] = useState<Valid>({
    name: { status: 'success', help: undefined },
    image: { status: 'success', help: undefined },
  });

  const [imagesSearchOptions, setImagesSearchOptions] = useState<string[]>();

  useEffect(() => {
    if (
      formTemplate.name &&
      formTemplate.image &&
      formTemplate.vmorcontainer &&
      valid.name.status === 'success' &&
      (template
        ? template.name !== formTemplate.name ||
          template.image !== formTemplate.image ||
          template.vmorcontainer !== formTemplate.vmorcontainer ||
          template.gui !== formTemplate.gui ||
          template.persistent !== formTemplate.persistent ||
          template.cpu !== formTemplate.cpu ||
          template.ram !== formTemplate.ram ||
          template.disk !== formTemplate.disk
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

  return (
    <Modal
      destroyOnClose={true}
      bodyStyle={{ paddingBottom: '5px' }}
      centered
      footer={null}
      title={template ? 'Modify template' : 'Create a new template'}
      visible={show}
      onCancel={closehandler}
      width="600px"
    >
      <Form
        labelCol={{ span: 2 }}
        wrapperCol={{ span: 22 }}
        form={form}
        onSubmitCapture={() => {
          submitHandler({
            ...formTemplate,
            image:
              images.find(i => getImageNoVer(i.name) === formTemplate.image)
                ?.name ?? formTemplate.image,
          })
            .then(() => {
              setShow(false);
              setFormTemplate(old => {
                return { ...old, name: undefined };
              });
              form.setFieldsValue({
                templatename: undefined,
              });
            })
            .catch(apolloErrorCatcher);
        }}
        initialValues={{
          templatename: formTemplate.name,
          image: formTemplate.image,
          vmorcontainer: formTemplate.vmorcontainer,
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

        <div className="flex justify-between items-start inline mb-6">
          <Form.Item
            className="my-0"
            {...fullLayout}
            style={{ width: '63%' }}
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
              options={imagesSearchOptions?.map(x => {
                return {
                  value: x,
                };
              })}
              onFocus={() => {
                if (!imagesSearchOptions?.length)
                  setImagesSearchOptions(imagesNoVersion!);
              }}
              onChange={value => {
                setImagesSearchOptions(
                  imagesNoVersion?.filter(s => s.includes(value))
                );
                if (value !== formTemplate.image) {
                  const imageFound = images.find(
                    i => getImageNoVer(i.name) === value
                  );
                  setFormTemplate(old => {
                    return {
                      ...old,
                      image: String(value),
                      registry: imageFound?.registry,
                      vmorcontainer:
                        imageFound?.vmorcontainer[0] ?? 'Container',
                      persistent: false,
                      gui: true,
                    };
                  });
                  form.setFieldsValue({
                    image: value,
                    vmorcontainer: imageFound?.vmorcontainer[0] ?? 'Container',
                  });
                }
              }}
              placeholder="Select an image"
            />
          </Form.Item>

          <div className="mt-3">
            <span>GUI:</span>
            <Checkbox
              className="ml-3"
              checked={formTemplate.gui}
              onChange={() =>
                setFormTemplate(old => {
                  return { ...old, gui: !old.gui };
                })
              }
            />
          </div>
          <div className="mr-1 mt-3">
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
          </div>
        </div>

        <Form.Item labelAlign="left" className="mt-10" label="CPU" name="cpu">
          <div className="sm:pl-3 pr-1">
            <Slider
              handleStyle={alternativeHandle}
              defaultValue={formTemplate.cpu}
              tooltipVisible={false}
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
              handleStyle={alternativeHandle}
              defaultValue={formTemplate.ram}
              tooltipVisible={false}
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
              handleStyle={alternativeHandle}
              tooltipVisible={false}
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
