import { Input, Form, Button, Row } from 'antd';
import { FC, useState } from 'react';
import { User } from '../UserList/UserList';

export interface IUserListFormProps {
  onAddUser: (newUser: User) => void;
  onCancel: () => void;
}

const UserListForm: FC<IUserListFormProps> = props => {
  const { onAddUser, onCancel } = props;
  const [validForm, setValidForm] = useState<boolean>(false);
  const [form] = Form.useForm();

  const cancelForm = () => {
    form.resetFields();
    onCancel();
  };

  // Used to disable Save button
  const handleChange = async (_: any, user: User) => {
    try {
      if (user.userid && user.name && user.email && user.surname) {
        await form.validateFields(['userid', 'name', 'surname', 'email']);
        setValidForm(true);
      } else {
        setValidForm(false);
        if (user.userid) await form.validateFields(['userid']);
        if (user.email) await form.validateFields(['email']);
        if (user.name) await form.validateFields(['name']);
        if (user.surname) await form.validateFields(['surname']);
      }
    } catch (_) {
      setValidForm(false);
    }
  };

  const submitForm = (user: User) => {
    onAddUser(user);
    form.resetFields();
  };

  const requiredFieldRule = {
    required: true,
    message: 'This field cannot be empty',
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
        name="userid"
        label="User ID"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        name="name"
        label="Name"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        name="surname"
        label="Surname"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input />
      </Form.Item>
      <Form.Item
        name="email"
        label="Email"
        validateTrigger="onBlur"
        rules={[
          {
            type: 'email',
            message: 'This is not a valid E-mail!',
          },

          requiredFieldRule,
        ]}
      >
        <Input type="email" />
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

export default UserListForm;
