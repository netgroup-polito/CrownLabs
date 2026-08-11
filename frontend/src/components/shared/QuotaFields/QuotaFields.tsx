import { Form, InputNumber, Space, Button, Select } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import type { FC, ReactNode } from 'react';
import type { Rule } from 'antd/es/form';

interface QuotaFieldRuleSet {
  cpu?: Rule[];
  memory?: Rule[];
  instances?: Rule[];
  disk?: Rule[];
  otherResources?: Rule[];
}

interface QuotaFieldLimit {
  min?: number;
  max?: number;
}

interface QuotaFieldLimitSet {
  cpu?: QuotaFieldLimit;
  memory?: QuotaFieldLimit;
  instances?: QuotaFieldLimit;
  disk?: QuotaFieldLimit;
  otherResources?: QuotaFieldLimit;
}

interface QuotaFieldTooltipSet {
  cpu?: ReactNode;
  memory?: ReactNode;
  instances?: ReactNode;
  disk?: ReactNode;
  otherResources?: ReactNode;
}

export interface IQuotaFieldsProps {
  disabled?: boolean;
  validateTrigger?: string;
  rules?: QuotaFieldRuleSet;
  limits?: QuotaFieldLimitSet;
  tooltips?: QuotaFieldTooltipSet;
}

const QuotaFields: FC<IQuotaFieldsProps> = ({
  disabled = false,
  validateTrigger,
  rules,
  limits,
  tooltips,
}) => {
  // Get the parent form instance from context safely
  const formInstance = Form.useFormInstance();

  // Watch the otherResources list values in real-time
  const currentOtherResources = Form.useWatch('otherResources') || [];

  // Extract all currently selected resource keys across all rows
  const selectedKeys = currentOtherResources
    .map((item: { key?: string }) => item?.key)
    .filter(Boolean);

  return (
    <>
      <Form.Item
        name="instances"
        label="Instances"
        validateTrigger={validateTrigger}
        rules={rules?.instances}
        tooltip={tooltips?.instances}
      >
        <InputNumber
          min={limits?.instances?.min}
          max={limits?.instances?.max}
          disabled={disabled}
          className="w-100"
        />
      </Form.Item>
      <Form.Item
        name="cpu"
        label="CPU"
        validateTrigger={validateTrigger}
        rules={rules?.cpu}
        tooltip={tooltips?.cpu}
      >
        <InputNumber
          min={limits?.cpu?.min}
          max={limits?.cpu?.max}
          disabled={disabled}
          className="w-100"
        />
      </Form.Item>

      <Form.Item
        name="memory"
        label="Memory (Gi)"
        validateTrigger={validateTrigger}
        rules={rules?.memory}
        tooltip={tooltips?.memory}
      >
        <InputNumber
          min={limits?.memory?.min}
          max={limits?.memory?.max}
          disabled={disabled}
          className="w-100"
          addonAfter="Gi"
        />
      </Form.Item>

      <Form.Item
        name="disk"
        label="Disk (Gi)"
        validateTrigger={validateTrigger}
        rules={rules?.disk}
        tooltip={tooltips?.disk}
      >
        <InputNumber
          min={limits?.disk?.min}
          max={limits?.disk?.max}
          disabled={disabled}
          className="w-100"
          addonAfter="Gi"
        />
      </Form.Item>

      {/* Extended Resources */}
      <Form.Item label="Extended Resources" tooltip={tooltips?.otherResources}>
        <Form.List name="otherResources">
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name, ...restField }) => {
                // Get the value selected in the current row to prevent it from self-disabling
                const currentFieldKey = formInstance?.getFieldValue([
                  'otherResources',
                  name,
                  'key',
                ]);

                return (
                  <Space
                    key={key}
                    style={{ display: 'flex', marginBottom: 8 }}
                    align="baseline"
                  >
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: 'Select a resource' }]}
                    >
                      <Select
                        placeholder="Resource Name"
                        style={{ width: 200 }}
                      >
                        <Select.Option
                          value="nvidiaComGpu"
                          disabled={
                            selectedKeys.includes('nvidiaComGpu') &&
                            currentFieldKey !== 'nvidiaComGpu'
                          }
                        >
                          Nvidia GPU
                        </Select.Option>
                        <Select.Option
                          value="amdComGpu"
                          disabled={
                            selectedKeys.includes('amdComGpu') &&
                            currentFieldKey !== 'amdComGpu'
                          }
                        >
                          AMD GPU
                        </Select.Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: 'Value required' }]}
                    >
                      <InputNumber placeholder="Amount" min={0} />
                    </Form.Item>
                    <MinusCircleOutlined onClick={() => remove(name)} />
                  </Space>
                );
              })}
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  style={{ width: 200 }}
                  icon={<PlusOutlined />}
                >
                  Add Resource
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Form.Item>
    </>
  );
};

export default QuotaFields;
