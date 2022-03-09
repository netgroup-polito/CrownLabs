/* eslint-disable react/no-multi-comp */
import { Table, Form, Popconfirm } from 'antd';
import {
  EditOutlined,
  CheckOutlined,
  CloseOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { UserAccountPage } from '../../../utils';
import { FC, useState } from 'react';
import EditableCell from './EditableCell';
import { SupportedError } from '../../../errorHandling/utils';
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
            user.key === updatedUser.key ? updatedUser : user
          )
        );
        props.setEditing(false);
      }
    } catch (errInfo) {
      props.genericErrorCatcher(errInfo as SupportedError);
    }
  };

  const columns = [
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
      render: (_: any, record: UserAccountPage) => {
        const editable = isEditing(record);
        return editable ? (
          <span>
            <CheckOutlined onClick={() => save(record)} className="mr-1" />

            <Popconfirm title="Sure to cancel?" onConfirm={cancel}>
              <CloseOutlined />
            </Popconfirm>
          </span>
        ) : (
          <>
            <EditOutlined
              disabled={editingKey !== ''}
              onClick={() => {
                edit(record);
              }}
            />
            <Popconfirm
              title="Sure to delete?"
              onConfirm={() =>
                props.updateUserCSV(
                  props.data.filter(user => user.key !== record.key)
                )
              }
            >
              <DeleteOutlined className="text-red-500" />
            </Popconfirm>
          </>
        );
      },
    },
  ];

  const makeOnCellCallback = (record: UserAccountPage, col: any) => ({
    record,
    inputType: 'text',
    dataIndex: col.dataIndex,
    title: col.title,
    editing: isEditing(record),
  });
  const mergedColumns = columns.map(col =>
    !col.editable
      ? col
      : {
          ...col,
          onCell: (record: UserAccountPage) => makeOnCellCallback(record, col),
        }
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
