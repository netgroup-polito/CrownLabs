import { Link } from 'react-router-dom';
import { Empty, Button, Table, Tooltip, Typography, Popconfirm } from 'antd';
import { calculateAge, compareAge } from '../../services/TimeUtils';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExportOutlined,
  LoadingOutlined
} from '@ant-design/icons';
import { getColumnSearchProps, ResizableTitle } from '../../services/TableUtils';
import React, { useEffect, useState } from 'react';

export default function Instances(props){

  const [columns, setColumns] = useState([]);

  useEffect(() => {
    setColumns([{
      dataIndex: 'Ready',
      key: 'Ready',
      title: 'Ready',
      width: '6em',
      fixed: true,
      align: 'center',
      sortDirections: ['descend', 'ascend'],
      sorter: {
        compare: (a, b) => a["Ready"] - b["Ready"],
      },
      render: text => text ?
        <Tooltip title={'VM Ready'}><CheckCircleOutlined style={{fontSize: 20, color: "#52c41a"}} /></Tooltip> :
        <Tooltip title={'Creating VM'}><LoadingOutlined style={{fontSize: 20}} /></Tooltip>
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
        ...getColumnSearchProps('Name', renderInstances, setColumns)
      },
      props.onProfessor ? {
        dataIndex: 'User',
        key: 'User',
        title: <div style={{marginLeft: '2em'}}>User</div>,
        sortDirections: ['descend', 'ascend'],
        sorter: {
          compare: (a, b) => a.User.localeCompare(b.User),
        },
        ...getColumnSearchProps('User', renderInstances, setColumns)
      } : {  },
      {
        dataIndex: 'IP',
        key: 'IP',
        title: <div style={{marginLeft: '2em'}}>IP</div>,
        ...getColumnSearchProps('URL', renderInstances, setColumns)
      },
      {
        title: 'Stop',
        key: 'Stop',
        width: '5em',
        align: 'center',
        render: (text, record) => {
          let lab = props.instances.find(lab => lab.metadata.name === record.key);
          return (
            <Popconfirm title={'Stop VM?'} onConfirm={() => window.api.deleteGenericResource(lab.metadata.selfLink)}>
              <Button icon={<CloseCircleOutlined style={{fontSize: 20, color: '#ff4d4f'}} />}
                      size={'small'} shape={'circle'}
                      style={{border: 'none', background: 'none'}}
              />
            </Popconfirm>
          )
        }
      },
      {
        title: 'Connect',
        key: 'Connect',
        width: '10em',
        align: 'center',
        render: (text, record) => {
          let lab = props.instances.find(lab => lab.metadata.name === record.key)
          let template = props.templates.find(item => item.metadata.name === lab.spec['template.crownlabs.polito.it/TemplateRef'].name)
          let url = (lab && lab.status) ? lab.status.url : '';
          return (
            (template && lab && template.spec.environmentList[0].guiEnabled && lab.status &&
              lab.status.phase === 'VmiReady') ? (
              <Tooltip title={'Connect VM'}>
                <a target={'_blank'} href={url}>
                  <Button icon={<ExportOutlined style={{fontSize: 20}} />}
                          size={'small'} shape={'circle'}
                          style={{border: 'none', background: 'none'}}
                  />
                </a>
              </Tooltip>
            ) : null
          )
        }
      },
      {
        dataIndex: 'Age',
        key: 'Age',
        title: 'Age',
        width: '5em',
        sorter: {
          compare: (a, b) => compareAge(a.Age, b.Age),
        },
        last: true
      }])
  }, [props.instances])

  const renderInstances = (text, record, dataIndex) => {
    return (
      dataIndex === 'Name' ? (
        <Typography.Text strong>{text}</Typography.Text>
      ) : (
        <div>{text}</div>
      )
    )
  }

  const instancesViews = [];
  props.instances.forEach(instances => {
    let template = props.templates.find(item => item.metadata.name === instances.spec['template.crownlabs.polito.it/TemplateRef'].name)
    if(template){
      instancesViews.push({
        key: instances.metadata.name,
        Ready: instances.status ? instances.status.phase === 'VmiReady' : false,
        Name: template.spec.description + ' - ' + instances.metadata.name.split('-').slice(-1),
        User: props.onProfessor ? instances.spec['tenant.crownlabs.polito.it/TenantRef'].name : null,
        IP: instances.status ? instances.status.ip : '',
        Age: calculateAge(instances.metadata.creationTimestamp)
      });
    }
  });

  return(
    <Table columns={columns} dataSource={instancesViews}
           pagination={{ position: ['bottomCenter'],
             hideOnSinglePage: instancesViews.length < 11,
             showSizeChanger: true,
           }} showSorterTooltip={false}
           components={{
             header: {
               cell: ResizableTitle
             }
           }}
           scroll={{ x: 'max-content' }}
           loading={props.loading}
           locale={{emptyText: <Empty description={'No Running Labs'} image={Empty.PRESENTED_IMAGE_SIMPLE} />}}
    />
  )
}
