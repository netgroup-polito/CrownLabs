import { Checkbox, Form, Slider, Space, Input, Tooltip, InputNumber } from 'antd';
import type { ChildFormItem, Interval, TemplateFormEnv } from './types';
import type { FC } from 'react';
import { formItemLayout } from './utils';

type EnvironmentDiskProps = {
  diskResources: Interval;
  isCloudVm?: boolean;
} & ChildFormItem;

export const EnvironmentDisk: FC<EnvironmentDiskProps> = ({
  parentFormName,
  restField,
  diskResources: resources,
  isCloudVm = false,
}) => {
  const form = Form.useFormInstance();

  const environments = Form.useWatch<TemplateFormEnv[] | undefined>(
    'environments',
  );

  const handlePersistentChange = (checked: boolean) => {
    if (!environments) return;
    if (!environments[parentFormName]) return;

    form.setFieldsValue({
      environments: environments.map((env, idx) => {
        if (idx === parentFormName) {
          return {
            ...env,
            persistent: checked,
            disk: checked ? resources.min : 0,
          };
        }
        return env;
      }),
    });
  };

  const handleDiskChange = (disk: number | null) => {
    if (disk === null) return;
    if (!environments) return;
    if (!environments[parentFormName]) return;

    form.setFieldsValue({
      environments: environments.map((env, idx) => {
        if (idx === parentFormName) {
          return {
            ...env,
            persistent: disk >= resources.min,
            disk: disk,
          };
        }
        return env;
      }),
    });
  };

  return (
    <Space direction='horizontal' className='ml-6 mt-4' >
    <Form.Item  className="mb-4" label="Disk" labelCol={{ span: 24 }}
        wrapperCol={{ span: 24 }}>
      <div className="flex gap-4 items-start">


        <div className="flex-1">
          <Form.Item {...restField} name={[parentFormName, 'disk']} noStyle className="mb-0">
            <Space.Compact block>
            <InputNumber
              max={resources.max}
              step={1}
              style={{ width: "30%", textAlignLast: "center" }}
              defaultValue={form.getFieldValue(['environments', parentFormName, 'disk']) || 0}
              onChange={value => handleDiskChange(value)}
            />
            <Input disabled className="site-input-right" value="DISK GB:" style={{
          width: "30%",
          borderInlineStart: 0,
          borderInlineEnd: 0,
          pointerEvents: 'none',
        }} />
            </Space.Compact>
          </Form.Item>
        </div>
        <Tooltip title="A persistent VM/container disk space won't be destroyed after being turned off.">
          <div className="pt-[6px]">
            <Form.Item
              {...restField}
              label="Persistent"
              name={[parentFormName, 'persistent']}
              valuePropName="checked"
              noStyle
            >
              <Checkbox
                disabled={isCloudVm}
                onChange={value => handlePersistentChange(value.target.checked)}>Persistent</Checkbox>
            </Form.Item>
          </div>
        </Tooltip>
      </div>
      
    </Form.Item>
    </Space> 
  );
};
