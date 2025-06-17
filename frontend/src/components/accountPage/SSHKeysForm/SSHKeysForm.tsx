import { Input, Form, Button, Row } from 'antd';
import type { FC } from 'react';
import { useState } from 'react';

export interface ISSHKeysFormProps {
  onSaveKey: (newKey: { name: string; key: string }) => void;
  onCancel: () => void;
}

const acceptedAlgorithms = [
  'sk-ecdsa-sha2-nistp256@openssh.com',
  'ecdsa-sha2-nistp256',
  'ecdsa-sha2-nistp384',
  'ecdsa-sha2-nistp521',
  'sk-ssh-ed25519@openssh.com',
  'ssh-ed25519',
  'ssh-dss',
  'ssh-rsa',
];

const SSHKeysForm: FC<ISSHKeysFormProps> = props => {
  const { onSaveKey, onCancel } = props;
  const [validForm, setValidForm] = useState<boolean>(false);
  const [form] = Form.useForm();

  const cancelForm = () => {
    form.resetFields();
    onCancel();
  };
  const validateSSHKey = async (_rules: unknown, key: string) => {
    if (!key) throw new Error('Key field is mandatory');
    const result = key.split(/\s+/); // split regardless the amount of spaces
    if (result.length !== 2 && result.length !== 3)
      throw new Error('Invalid key format');
    if (result[1] === '') throw new Error('Empty SSH key');
    if (!acceptedAlgorithms.includes(result[0]))
      throw new Error('Unrecognized SSH key format');
  };

  // Used to disable Save button
  const handleChange = async (
    _: unknown,
    { name, key }: { name: string; key: string },
  ) => {
    try {
      if (key && name) {
        await form.validateFields(['name', 'key']);
        setValidForm(true);
      } else if (key) await form.validateFields(['key']);
      else await form.validateFields(['name']);
    } catch (_) {
      setValidForm(false);
    }
  };
  const submitForm = ({ name, key }: { name: string; key: string }) => {
    // the name field is appended to the key
    const comment = key.split(/\s+/)[2];
    name = name.trim();
    if (comment) key = `${key}:${name}`;
    else key = `${key} ${name}`;

    onSaveKey({ name, key });
    form.resetFields();
  };

  return (
    <Form
      form={form}
      labelCol={{ span: 4 }}
      wrapperCol={{ span: 24 }}
      onFinish={submitForm}
      onValuesChange={handleChange}
    >
      <Form.Item
        name="name"
        label="Name"
        validateTrigger="onBlur"
        rules={[
          {
            required: true,
            message: 'Field required',
          },
        ]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        validateTrigger="onBlur"
        name="key"
        label="Public Key"
        rules={[
          {
            required: true,
            validator: validateSSHKey,
          },
        ]}
      >
        <Input.TextArea
          rows={10}
          placeholder={`Your PUBLIC KEY here \n Begins with ${acceptedAlgorithms
            .map(el => "'" + el + "'")
            .join(', ')}`}
        />
      </Form.Item>
      <Form.Item>
        <Row justify="end">
          <Button type="default" htmlType="button" onClick={cancelForm}>
            Cancel
          </Button>
          <Button
            disabled={!validForm}
            className="ml-2"
            type="primary"
            htmlType="submit"
          >
            Save
          </Button>
        </Row>
      </Form.Item>
    </Form>
  );
};

export default SSHKeysForm;
