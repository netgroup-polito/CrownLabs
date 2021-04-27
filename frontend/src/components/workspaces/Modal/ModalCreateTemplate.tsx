import { FC } from 'react';
import { Modal, Slider, Form, Input } from 'antd';
import Button from 'antd-button-color';
import '../../../index.less'; //To delete, usefull only to storybook

export interface IModalCreateTemplateProps {
  showmodal: boolean;
  setshowmodal: (status: boolean) => void;
}

const ModalCreateTemplate: FC<IModalCreateTemplateProps> = ({ ...props }) => {
  const { showmodal, setshowmodal } = props;

  const handleCancel = () => {
    setshowmodal(false);
  };

  const formatterCPU = (value?: number) => {
    return `${value} CPUs`;
  };

  const formatterRAM = (value?: number) => {
    return `${value} GB`;
  };

  const formatterDISK = (value?: number) => {
    return `${value} GB`;
  };

  const layout = {
    labelCol: { span: 6 },
    wrapperCol: { span: 18 },
  };
  const tailLayout = {
    wrapperCol: { offset: 0, span: 24 },
  };

  const Center = (
    <Form {...layout}>
      <Form.Item label="Template Name">
        <div className="pl-4 pr-2">
          <Input />
        </div>
      </Form.Item>
      <Form.Item label="CPU" name="cpu">
        <div className="pl-3">
          <Slider
            defaultValue={0}
            min={1}
            max={4}
            step={1}
            tipFormatter={formatterCPU}
          />
        </div>
      </Form.Item>
      <Form.Item label="RAM" name="ram">
        <div className="pl-3">
          <Slider
            defaultValue={0}
            min={1}
            max={8}
            step={1}
            tipFormatter={formatterRAM}
          />
        </div>
      </Form.Item>
      <Form.Item label="DISK" name="disk">
        <Slider
          className="ml-4"
          defaultValue={0}
          min={8}
          max={16}
          step={1}
          tipFormatter={formatterDISK}
        />
      </Form.Item>
      <Form.Item {...tailLayout}>
        <div className="flex justify-center">
          <Button
            htmlType="submit"
            type="primary"
            shape="round"
            size={'middle'}
            onClick={() => setshowmodal(false)}
          >
            Create
          </Button>
        </div>
      </Form.Item>
    </Form>
  );

  return (
    <Modal
      bodyStyle={{ paddingBottom: '5px' }}
      centered
      footer={null}
      title="Create a new template"
      visible={showmodal}
      onCancel={handleCancel}
    >
      {Center}
    </Modal>
  );
};

export default ModalCreateTemplate;
