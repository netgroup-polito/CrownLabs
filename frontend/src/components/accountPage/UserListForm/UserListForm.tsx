import { Input, Form, Button, Row, AutoComplete, Select } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import type { FC } from 'react';
import { useState } from 'react';
import type { UserAccountPage } from '../../../utils';
import { filterUser } from '../../../utils';
import { Role } from '../../../generated-types';
export interface IUserListFormProps {
  onAddUser: (newUser: UserAccountPage, role: Role) => void;
  onCancel: () => void;
  users: UserAccountPage[];
}

const { Option } = AutoComplete;

const UserListForm: FC<IUserListFormProps> = props => {
  const { onAddUser, onCancel, users } = props;
  const [validForm, setValidForm] = useState<boolean>(false);
  const [searched, setSearched] = useState<boolean>(false);
  const [searchedText, setSearchedText] = useState<string>('');
  const [role, setRole] = useState<Role>(Role.User);
  const [form] = Form.useForm();

  const cancelForm = () => {
    form.resetFields();
    onCancel();
  };

  // Used to disable Save button
  const handleChange = async (_: unknown, _user: UserAccountPage) => {
    try {
      await form.validateFields(['userid']);
      await form.validateFields(['email']);
      await form.validateFields(['name']);
      await form.validateFields(['surname']);
      setValidForm(true);
    } catch (e) {
      console.error('Form validation failed:', e);
      setValidForm(false);
    }
  };

  const submitForm = (user: UserAccountPage) => {
    const { ...newUser } = user;
    onAddUser(newUser, role);
    setSearched(false);
    form.resetFields();
  };

  const requiredFieldRule = {
    required: true,
    message: 'This field cannot be empty',
  };

  const handleSearch = (value: string) => {
    setSearchedText(value);

    if (value === '') {
      setSearched(false);
      form.resetFields();
    }
  };

  const handleSelect = (value: string) => {
    setSearched(true);
    const selectedUser: UserAccountPage = users.filter(
      user => user.userid === value,
    )[0];

    form.setFieldsValue({
      searchInput: searchedText,
      userid: selectedUser.userid,
      name: selectedUser.name,
      surname: selectedUser.surname,
      email: selectedUser.email,
      role: role,
    });
    handleChange(form, selectedUser);
  };

  return (
    <Form
      form={form}
      labelCol={{ span: 4 }}
      wrapperCol={{ span: 24 }}
      onFinish={submitForm}
      onValuesChange={handleChange}
    >
      <Form.Item name="searchInput" label={<SearchOutlined />}>
        <AutoComplete
          onSearch={handleSearch}
          onSelect={handleSelect}
          placeholder="Search user"
          allowClear={true}
        >
          {searchedText === ''
            ? users.map((user, key) => (
                <Option key={key} value={user.userid}>
                  <i className="mr-2"> {user.userid}</i>&ndash;
                  <span className="ml-2">
                    {user.name} {user.surname}
                  </span>
                </Option>
              ))
            : users
                .filter(user => filterUser(user, searchedText))
                .map((user, key) => (
                  <Option key={key} value={user.userid}>
                    <p>
                      {user.userid} | {user.name} {user.surname}
                    </p>
                  </Option>
                ))}
        </AutoComplete>
      </Form.Item>
      <Form.Item
        name="userid"
        label="User ID"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input disabled={searched} />
      </Form.Item>
      <Form.Item
        name="name"
        label="Name"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input disabled={searched} />
      </Form.Item>
      <Form.Item
        name="surname"
        label="Surname"
        validateTrigger="onBlur"
        rules={[requiredFieldRule]}
      >
        <Input disabled={searched} />
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
        <Input type="email" disabled={searched} />
      </Form.Item>
      <Form.Item name="role" label="Role">
        <Select defaultValue={Role.User} onChange={value => setRole(value)}>
          <Select.Option value={Role.User}> User </Select.Option>
          <Select.Option value={Role.Manager}> Manager </Select.Option>
        </Select>
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
