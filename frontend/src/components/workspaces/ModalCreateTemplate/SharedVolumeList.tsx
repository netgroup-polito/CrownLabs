import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Checkbox, Form, Input, Select, Tooltip } from 'antd';
import type { FC } from 'react';
import type { SharedVolume } from '../../../utils';

const readonlyTooltipTitle = `When checked, the selected Shared Volume will be mounted with Read-Only permissions. Keep in mind that expert users may edit this setting if mounted on a VM.`;

const fullLayout = {
  wrapperCol: { offset: 0, span: 24 },
};

type SharedVolumeListProps = {
  parentFormName: number;
  sharedVolumes: SharedVolume[];
};

export const SharedVolumeList: FC<SharedVolumeListProps> = ({
  parentFormName,
  sharedVolumes,
}) => {
  const validateMouthPath = async (_: unknown, value: string) => {
    if (!value || value.trim() === '') {
      throw new Error('Mount Path cannot be empty');
    }
    if (!/^\/([a-zA-Z0-9_-]+\/?)*$/.test(value)) {
      throw new Error('Mount Path is not valid');
    }
  };

  return (
    <Form.List name={[parentFormName, 'sharedVolumeMounts']}>
      {(fields, { add, remove }) => (
        <>
          {fields.map(({ key, name, ...restField }) => (
            <div key={key} className="flex items-baseline gap-4">
              <div className="flex items-baseline gap-2 flex-1">
                <Form.Item
                  {...restField}
                  name={[name, 'sharedVolume']}
                  rules={[{ required: true, message: 'Missing Shared Volume' }]}
                   getValueFromEvent={(value) => {
                    const selected = sharedVolumes.find(sv => sv.id === value);
                    return selected?.name ?? value;
                  }}
                  getValueProps={(value) => {
                    const selected = sharedVolumes.find(sv => sv.name === value);
                    return { value: selected?.id ?? value };
                  }}
                  
                >
                  <Select placeholder="Select..." style={{ width: '160px' }}>
                    {sharedVolumes.map(shvol => (
                      <Select.Option key={shvol.id} value={shvol.id}>
                        {shvol.prettyName}
                      </Select.Option>
                    ))}
                  </Select>
                </Form.Item>

                <Form.Item
                  {...restField}
                  name={[name, 'mountPath']}
                  rules={[
                    { required: true, message: 'Missing Mount Path' },
                    {
                      validator: validateMouthPath,
                    },
                  ]}
                  validateDebounce={500}
                  className="flex-1"
                >
                  <Input placeholder="Example: /mnt/myvol" />
                </Form.Item>

                <Tooltip title={readonlyTooltipTitle}>
                  <Form.Item
                    {...restField}
                    label="RO"
                    name={[name, 'readOnly']}
                    valuePropName="checked"
                  >
                    <Checkbox />
                  </Form.Item>
                </Tooltip>
              </div>

              <div className="text-red-500">
                <MinusCircleOutlined onClick={() => remove(name)} />
              </div>
            </div>
          ))}

          <Form.Item {...fullLayout}>
            <Button
              type="dashed"
              onClick={() =>
                add({
                  sharedVolumes: '',
                  mountPath: '',
                  readOnly: false,
                })
              }
              block
              icon={<PlusOutlined />}
            >
              Mount a Shared Volume
            </Button>
          </Form.Item>
        </>
      )}
    </Form.List>
  );
};
