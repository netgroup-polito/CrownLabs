import { Checkbox, Form, Slider, Tooltip } from 'antd';
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

  const handleDiskChange = (disk: number) => {
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
    <Form.Item label="Persistent" {...formItemLayout}>
      <div className="flex gap-4 items-start">
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
                onChange={value => handlePersistentChange(value.target.checked)}
              />
            </Form.Item>
          </div>
        </Tooltip>

        <div className="flex-1">
          <Form.Item {...restField} name={[parentFormName, 'disk']} noStyle>
            <Slider
              tooltip={{
                defaultOpen: false,
                formatter: value => `${value} GB`,
              }}
              max={resources.max}
              marks={{
                0: '0GB',
                [resources.min]: `${resources.min}GB`,
                [resources.max]: `${resources.max}GB`,
              }}
              onChangeComplete={value => handleDiskChange(value)}
            />
          </Form.Item>
        </div>
      </div>
    </Form.Item>
  );
};
