import { Button, Form, Input, Row } from 'antd';
import type { FC } from 'react';

export interface ITenantSearchFormProps {
  onSearch: (tenantId: string) => void;
  isLoading?: boolean;
}

const TenantSearchForm: FC<ITenantSearchFormProps> = ({
  onSearch,
  isLoading = false,
}) => {
  const [form] = Form.useForm();

  const submitForm = ({ tenantId }: { tenantId: string }) => {
    onSearch(tenantId.trim().toLowerCase());
  };

  return (
    <div className="w-full max-w-lg">
      <h2 className="md:text-2xl text-lg text-center mb-4">Search Tenant</h2>

      <Form
        form={form}
        labelCol={{ span: 4 }}
        wrapperCol={{ span: 24 }}
        onFinish={submitForm}
      >
        <Form.Item
          name="tenantId"
          label="Tenant ID"
          validateTrigger="onBlur"
          rules={[
            {
              required: true,
              message: 'ID required',
            },
          ]}
        >
          <Input />
        </Form.Item>

        <Row justify="center">
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={isLoading === true}
            >
              Search
            </Button>
          </Form.Item>
        </Row>
      </Form>
    </div>
  );
};

export default TenantSearchForm;
