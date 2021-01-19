import { Button, Empty, Popconfirm, Table, Tooltip, Typography } from 'antd';
import React, { useEffect, useState } from 'react';
import {
  CodeOutlined,
  DeleteOutlined,
  DesktopOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { getColumnSearchProps, ResizableTitle } from '../../services/TableUtils';
import Utils from '../../services/Utils';

export default function Templates(props){

  const [columns, setColumns] = useState([]);

  useEffect(() => {
    setColumns([{
      dataIndex: 'VM Type',
      key: 'VM Type',
      title: 'Type',
      width: '5em',
      align: 'center',
      sortDirections: ['descend', 'ascend'],
      sorter: {
        compare: (a, b) => a["VM Type"] - b["VM Type"],
      },
      render: text => text ?
        <Tooltip title={'GUI enabled'}><DesktopOutlined style={{fontSize: 20}} /></Tooltip> :
        <Tooltip title={'CLI only'}><CodeOutlined style={{fontSize: 20}} /></Tooltip>
    },
      {
        dataIndex: 'Name',
        key: 'Name',
        title: <div style={{marginLeft: '2em'}}>Name</div>,
        sortDirections: ['descend', 'ascend'],
        defaultSortOrder: 'ascend',
        sorter: {
          compare: (a, b) => a.Name.localeCompare(b.Name),
        },
        ...getColumnSearchProps('Name', renderTemplates, setColumns)
      },
      {
        dataIndex: 'Course',
        key: 'Course',
        title: <div style={{marginLeft: '2em'}}>Course</div>,
        sortDirections: ['descend', 'ascend'],
        sorter: {
          compare: (a, b) => a.Course.localeCompare(b.Course),
        },
        ...getColumnSearchProps('Course', renderTemplates, setColumns)
      },
      props.onProfessor ? {
        dataIndex: 'Delete',
        title: 'Delete',
        key: 'Delete',
        width: '6em',
        align: 'center',
        render: (text, record) => {
          const lab = props.templates.find(lab => lab.metadata.name === record.key);

          return (
            <Popconfirm title={'Delete Lab?'} onConfirm={() => window.api.deleteGenericResource(lab.metadata.selfLink)}>
              <Button icon={<DeleteOutlined style={{fontSize: 20, color: '#ff4d4f'}}  />}
                      size={'small'} shape={'circle'}
                      style={{border: 'none', background: 'none'}}
              />
            </Popconfirm>
          )
        },
      } : {  },
      {
        title: 'Start',
        key: 'Start',
        width: '5em',
        align: 'center',
        render: (text, record) => (
          <Tooltip title={'Create VM'}>
            <Button icon={<PlayCircleOutlined style={{fontSize: 20, color: '#1890ff'}} />}
                    size={'small'} shape={'circle'}
                    style={{border: 'none', background: 'none'}}
                    onClick={() => startLab(props.templates.find(lab => lab.metadata.name === record.key))}
            />
          </Tooltip>
        ),
      }])
  }, [props.templates])

  const renderTemplates = (text, record, dataIndex) => {
    return (
      dataIndex === 'Name' ? (
        <Typography.Text strong>{text}</Typography.Text>
      ) : (
        <div>{text}</div>
      )
    )
  }

  const templatesViews = [];
  props.templates.forEach(templates => {
    templatesViews.push({
      key: templates.metadata.name,
      "VM Type": templates.spec.environmentList[0].guiEnabled,
      Name: templates.spec.prettyName,
      Course: templates.spec['workspace.crownlabs.polito.it/WorkspaceRef'].name
    });
  });

  const startLab = template => {
    const templatesName = template.metadata.name;
    const templatesNamespace = template.metadata.namespace;

    let studentID = Utils().parseJWT().preferred_username;
    let instanceNamespace = props.tenants[0].status.personalNamespace.name;

    let item = {
      spec: {
        ['template.crownlabs.polito.it/TemplateRef']: {
          name: templatesName,
          namespace: templatesNamespace,
        },
        ['tenant.crownlabs.polito.it/TenantRef']: {
          name: studentID
        }
      },
      metadata: {
        name: templatesName + '-' + studentID + '-' +
          Math.floor(Math.random() * 1000) + 1,
        namespace: instanceNamespace
      },
      apiVersion: 'crownlabs.polito.it/v1alpha2',
      kind: 'Instance'
    }

    window.api.createCustomResource(
      'crownlabs.polito.it',
      'v1alpha2',
      instanceNamespace,
      'instances',
      item
    ).catch(error => console.log(error));
  }

  return(
    <Table columns={columns} dataSource={templatesViews}
           pagination={{ position: ['bottomCenter'],
             hideOnSinglePage: templatesViews.length < 11,
             showSizeChanger: true,
           }} showSorterTooltip={false}
           components={{
             header: {
               cell: ResizableTitle
             }
           }}
           scroll={{ x: 'max-content' }}
           loading={props.loading}
           locale={{emptyText: <Empty description={'No Available Labs'} image={Empty.PRESENTED_IMAGE_SIMPLE} />}}
    />
  )
}
