import { useState, useEffect, FC } from 'react';
import {
  Modal,
  Slider,
  Form,
  Input,
  Select,
  Checkbox,
  Tooltip,
  Row,
} from 'antd';
import Button from 'antd-button-color';

const alternativeHandle = { border: 'solid 2px #1c7afdd8' };

const { Option } = Select;

type Image = {
  name: string;
  vmorcontainer: Array<Vmorcontainer>;
};

type Vmorcontainer = 'Container' | 'VM';

type Template = {
  name: string | undefined;
  image: string | undefined;
  vmorcontainer: Vmorcontainer | undefined;
  diskMode: boolean;
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
  name: { status: string; help: string | undefined };
  vmorcontainer: { status: string; help: string | undefined };
};
export interface IModalCreateTemplateProps {
  template?: Template;
  images: Array<Image>;
  cpuInterval: Interval;
  ramInterval: Interval;
  diskInterval: Interval;
  show: boolean;
  setShow: (status: boolean) => void;
  submitHandler: (t?: Template) => void;
}

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
  } = props;

  const [buttonDisabled, setButtonDisabled] = useState(true);

  const [formTemplate, setFormTemplate] = useState<Template>({
    name: template && template.name,
    image: template && template.image,
    vmorcontainer: template && template.vmorcontainer,
    diskMode: !!template && template.diskMode,
    gui: !!template && template.gui,
    cpu: template ? template.cpu : cpuInterval.min,
    ram: template ? template.ram : ramInterval.min,
    disk: template ? template.disk : diskInterval.min,
  });

  const [valid, setValid] = useState<Valid>({
    name: { status: 'success', help: undefined },
    vmorcontainer: { status: 'success', help: undefined },
  });

  useEffect(() => {
    if (
      formTemplate.name &&
      formTemplate.image &&
      formTemplate.vmorcontainer &&
      (template
        ? template.name !== formTemplate.name ||
          template.image !== formTemplate.image ||
          template.vmorcontainer !== formTemplate.vmorcontainer ||
          template.gui !== formTemplate.gui ||
          template.diskMode !== formTemplate.diskMode ||
          template.cpu !== formTemplate.cpu ||
          template.ram !== formTemplate.ram ||
          template.disk !== formTemplate.disk
        : true)
    )
      setButtonDisabled(false);
    else setButtonDisabled(true);
  }, [formTemplate, template]);

  const nameValidator = () => {
    if (formTemplate.name === '' || formTemplate.name === undefined) {
      setValid(old => {
        return {
          ...old,
          name: { status: 'error', help: 'Please insert template name' },
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

  const vmorcontainerValidator = () => {
    if (formTemplate.vmorcontainer === undefined) {
      setValid(old => {
        return {
          ...old,
          vmorcontainer: { status: 'error', help: 'Please select' },
        };
      });
    } else {
      setValid(old => {
        return {
          ...old,
          vmorcontainer: { status: 'success', help: undefined },
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
        onSubmitCapture={() => submitHandler(formTemplate)}
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
          validateTrigger="onBlur"
          rules={[
            {
              required: true,
              validator: nameValidator,
            },
          ]}
        >
          <Input
            onBlur={e =>
              setFormTemplate(old => {
                return { ...old, name: e.target.value };
              })
            }
            placeholder="Insert template name"
            allowClear
          />
        </Form.Item>

        <div className="flex justify-between inline">
          <Form.Item
            {...fullLayout}
            style={{ width: '68%' }}
            name="image"
            required
          >
            <Select
              placeholder="Image name"
              onSelect={value => {
                if (value !== formTemplate.image) {
                  setFormTemplate(old => {
                    if (old.image) {
                      setValid(old => {
                        return {
                          ...old,
                          vmorcontainer: {
                            status: 'error',
                            help: 'Please select',
                          },
                        };
                      });
                    }
                    return {
                      ...old,
                      image: String(value),
                      vmorcontainer: undefined,
                      diskMode: false,
                      gui: false,
                    };
                  });
                  form.setFieldsValue({
                    image: value,
                    vmorcontainer: undefined,
                  });
                }
              }}
              showSearch={true}
            >
              {images.map(x => (
                <Option value={x.name} key={x.name}>
                  {x.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            {...fullLayout}
            style={{ width: '28%' }}
            name="vmorcontainer"
            required
            validateStatus={valid.vmorcontainer.status as 'success' | 'error'}
            help={valid.vmorcontainer.help}
            rules={[
              {
                required: true,
                validator: vmorcontainerValidator,
              },
            ]}
          >
            <Select
              placeholder="VM/Container"
              onSelect={value => {
                if (value !== formTemplate.vmorcontainer) {
                  setFormTemplate(old => {
                    return {
                      ...old,
                      vmorcontainer: String(value) as Vmorcontainer,
                      diskMode: false,
                      gui: false,
                      disk: 0,
                    };
                  });
                  form.setFieldsValue({
                    vmorcontainer: value,
                  });
                }
              }}
              showSearch={true}
            >
              {formTemplate.image !== undefined
                ? images
                    .filter(x => x.name === formTemplate.image)[0]
                    .vmorcontainer.map(x => (
                      <Option value={x} key={x}>
                        {x}
                      </Option>
                    ))
                : null}
            </Select>
          </Form.Item>
        </div>
        <Row className="flex mb-8 justify-center">
          <div className="mr-8 md:mr-12">
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
          <div className="ml-8 md:ml-12">
            <span>Persistent: </span>
            <Tooltip title="A persistent VM/container disk space won't be destroyed after being turned off.">
              <Checkbox
                className="ml-3"
                checked={formTemplate.diskMode}
                onChange={() =>
                  setFormTemplate(old => {
                    return {
                      ...old,
                      diskMode: !old.diskMode,
                      disk: !old.diskMode
                        ? template?.disk || diskInterval.min
                        : 0,
                    };
                  })
                }
              />
            </Tooltip>
          </div>
        </Row>

        <Form.Item labelAlign="left" className="mt-3" label="CPU" name="cpu">
          <div className="sm:px-3">
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
          <div className="sm:px-3">
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
              step={1}
              tipFormatter={(value?: number) => `${value} GB`}
            />
          </div>
        </Form.Item>
        <Form.Item
          labelAlign="left"
          label="DISK"
          name="disk"
          className={formTemplate.diskMode ? '' : 'hidden'}
        >
          <div className="sm:px-3 ">
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
                    className="pointer-events-none"
                    disabled
                    htmlType="submit"
                    type="primary"
                    shape="round"
                    size={'middle'}
                  >
                    {template ? 'Modify' : 'Create'}
                  </Button>
                </span>
              </Tooltip>
            ) : (
              <Button
                htmlType="submit"
                type="primary"
                shape="round"
                size={'middle'}
              >
                {template ? 'Modify' : 'Create'}
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
