import type { FC } from 'react';
import { Input, Form } from 'antd';
import type { UserAccountPage } from '../../../utils';
import type React from 'react';

export interface EditableCellProps {
  editing: boolean;
  dataIndex: string;
  title: React.ReactNode;
  inputType: 'number' | 'text';
  record: UserAccountPage;
  index: number;
  children: React.ReactNode;
}

const EditableCell: FC<EditableCellProps> = ({
  editing,
  dataIndex,
  title,
  children,
  ...restProps
}) => {
  const inputNode = <Input />;

  return (
    <td {...restProps}>
      {editing ? (
        <Form.Item
          name={dataIndex}
          style={{ margin: 0 }}
          rules={[
            {
              required: true,
              message: `Please Input ${title}!`,
            },
          ]}
        >
          {inputNode}
        </Form.Item>
      ) : (
        children
      )}
    </td>
  );
};

export default EditableCell;
