import { Table, Form, Popconfirm, Tooltip } from 'antd';
import {
  EditOutlined,
  CheckOutlined,
  CloseOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import type { UserAccountPage } from '../../../utils';
import type { FC } from 'react';
import { useState } from 'react';
import EditableCell from './EditableCell';
import type { SupportedError } from '../../../errorHandling/utils';
import type { ColumnType } from 'antd/lib/table';
export interface IEditableTableProps {
  data: UserAccountPage[];
  updateUserCSV: (user: UserAccountPage[]) => void;
  setEditing: (value: boolean) => void;
  genericErrorCatcher: (err: SupportedError) => void;
}

const EditableTable: FC<IEditableTableProps> = props => {
  const [form] = Form.useForm();
  const [editingKey, setEditingKey] = useState('');

  const isEditing = (record: UserAccountPage) => record.key === editingKey;

  const edit = (record: Partial<UserAccountPage> & { key: React.Key }) => {
    form.setFieldsValue({
      userid: '',
      name: '',
      surname: '',
      email: '',
      ...record,
    });
    setEditingKey(record.key);
    props.setEditing(true);
  };

  const cancel = () => {
    setEditingKey('');
    props.setEditing(false);
  };

  const save = async (record: UserAccountPage) => {
    try {
      const row = (await form.validateFields()) as UserAccountPage;

      const item = props.data.find(item => record.key === item.key);
      if (item) {
        const updatedUser = {
          ...item,
          ...row,
        };

        setEditingKey('');
        props.updateUserCSV(
          props.data.map(user =>
            user.key === updatedUser.key ? updatedUser : user,
          ),
        );
        props.setEditing(false);
      }
    } catch (errInfo) {
      props.genericErrorCatcher(errInfo as SupportedError);
    }
  };

  const columns: (ColumnType<UserAccountPage> & { editable?: boolean })[] = [
    {
      title: 'User ID',
      key: 'userid',
      dataIndex: 'userid',
      width: '25%',
    },

    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: '25%',
      editable: true,
    },
    {
      title: 'Surname',
      dataIndex: 'surname',
      key: 'surname',
      width: '15%',
      editable: true,
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      width: '40%',
      editable: true,
    },
    {
      title: 'Action',
      dataIndex: 'operation',
      render: (_: unknown, record: UserAccountPage, _index: number) => {
        const editable = isEditing(record);
        return editable ? (
          <span>
            <Tooltip title="Confirm changes">
              <CheckOutlined onClick={() => save(record)} className="mr-1" />
            </Tooltip>

            <Popconfirm title="Sure to cancel?" onConfirm={cancel}>
              <Tooltip title="Discard changes">
                <CloseOutlined />
              </Tooltip>
            </Popconfirm>
          </span>
        ) : (
          <>
            <Tooltip title="Edit details">
              <EditOutlined
                className="mx-1"
                disabled={editingKey !== ''}
                onClick={() => edit(record)}
              />
            </Tooltip>
            <Popconfirm
              title="Sure to delete?"
              onConfirm={() =>
                props.updateUserCSV(
                  props.data.filter(user => user.key !== record.key),
                )
              }
            >
              <Tooltip title="Remove entry">
                <DeleteOutlined className="text-red-500" />
              </Tooltip>
            </Popconfirm>
          </>
        );
      },
    },
  ];

  const makeOnCellCallback = (
    record: UserAccountPage,
    col: ColumnType<UserAccountPage>,
  ) => ({
    record,
    inputType: 'text',
    dataIndex: col.dataIndex,
    title: typeof col.title === 'string' ? col.title : undefined,
    editing: isEditing(record),
  });

  const mergedColumns: ColumnType<UserAccountPage>[] = columns.map(col =>
    !col.editable
      ? col
      : {
          ...col,
          onCell: (record: UserAccountPage) => makeOnCellCallback(record, col),
        },
  );

  return (
    <Form form={form} component={false}>
      <Table
        pagination={{ defaultPageSize: 10, size: 'small' }}
        components={{
          body: {
            cell: EditableCell,
          },
        }}
        size="small"
        dataSource={props.data}
        columns={mergedColumns}
        rowClassName="editable-row"
      />
    </Form>
  );
};

export default EditableTable;
