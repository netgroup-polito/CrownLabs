import { Link } from 'react-router-dom';
import { Empty, Button, Table, Tooltip, Typography, Popconfirm } from 'antd';
import { calculateAge, compareAge } from '../../services/TimeUtils';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  CodeOutlined,
  PlayCircleOutlined,
  DesktopOutlined,
  PauseCircleOutlined,
  DeleteOutlined,
  ExportOutlined,
  LoadingOutlined,
  FolderOpenOutlined,
} from '@ant-design/icons';
import {
  getColumnSearchProps,
  ResizableTitle
} from '../../services/TableUtils';
import React, { useEffect, useState } from 'react';
import { colorBlue } from './Templates';
const colorGreen = '#52c41a';
const colorRed = '#ff4d4f';
const colorYellow = '#e8be15';

export default function Instances(props) {
  const [columns, setColumns] = useState([]);
  const [terminating, setTerminating] = useState([]);

  useEffect(() => {
    setColumns([
      {
        dataIndex: 'Status',
        key: 'Status',
        title: 'Status',
        width: '6em',
        fixed: true,
        align: 'center',
        sortDirections: ['descend', 'ascend'],
        sorter: {
          compare: (a, b) => a['Status'] - b['Status']
        },
        render: (phase, record) => {
          let terminate = !!terminating.find(i => i === record.key);
          let lab = props.instances.find(
            lab => lab.metadata.name === record.key
          );
          const renderReadyIcon = (terminating, phase) => {
            const loading = (
              <Tooltip title={'Changing VM status'}>
                <LoadingOutlined style={{ fontSize: 20 }} />
              </Tooltip>
            );
            if (terminating) return loading;
            switch (phase) {
              case 'VmiReady':
                return (
                  <Tooltip title={'VM Ready'}>
                    <CheckCircleOutlined
                      style={{ fontSize: 20, color: colorGreen }}
                    />
                  </Tooltip>
                );
              case 'VmiOff':
                return (
                  <Tooltip title={'VM off'}>
                    <PauseCircleOutlined
                      style={{ fontSize: 20, color: colorYellow }}
                    />
                  </Tooltip>
                );
              default:
                return loading;
            }
          };
          return {
            children: renderReadyIcon(terminate, phase),
            props: {
              title: ''
            }
          };
        }
      },
      {
        dataIndex: 'VM Type',
        key: 'VM Type',
        title: 'Type',
        width: '5em',
        align: 'center',
        sortDirections: ['descend', 'ascend'],
        sorter: {
          compare: (a, b) => a['VM Type'] - b['VM Type']
        },
        render: text => {
          return {
            children: text ? (
              <Tooltip title={'GUI enabled'}>
                <DesktopOutlined style={{ fontSize: 20 }} />
              </Tooltip>
            ) : (
              <Tooltip title={'CLI only'}>
                <CodeOutlined style={{ fontSize: 20 }} />
              </Tooltip>
            ),
            props: {
              title: ''
            }
          };
        }
      },
      {
        dataIndex: 'Name',
        key: 'Name',
        title: <div style={{ marginLeft: '2em' }}>Name</div>,
        sortDirections: ['descend', 'ascend'],
        defaultSortOrder: 'ascend',
        sorter: {
          compare: (a, b) => a.Name.localeCompare(b.Name)
        },
        ...getColumnSearchProps('Name', renderInstances, setColumns)
      },
      props.onProfessor
        ? {
            dataIndex: 'User',
            key: 'User',
            title: <div style={{ marginLeft: '2em' }}>User</div>,
            sortDirections: ['descend', 'ascend'],
            sorter: {
              compare: (a, b) => a.User.localeCompare(b.User)
            },
            ...getColumnSearchProps('User', renderInstances, setColumns)
          }
        : {},
      {
        dataIndex: 'IP',
        key: 'IP',
        title: <div style={{ marginLeft: '2em' }}>IP</div>,
        ...getColumnSearchProps('URL', renderInstances, setColumns)
      },
      {
        title: 'Connect',
        key: 'Connect',
        width: '7em',
        align: 'center',
        render: (text, record) => {
          let lab = props.instances.find(
            lab => lab.metadata.name === record.key
          );
          if (lab) {
            let template = props.templates.find(
              item =>
                item.metadata.name ===
                lab.spec['template.crownlabs.polito.it/TemplateRef'].name
            );
            let url = lab && lab.status ? lab.status.url : '';
            return {
              children:
                template &&
                template.spec.environmentList[0].guiEnabled &&
                lab.status &&
                lab.status.phase === 'VmiReady' ? (
                  <a target={'_blank'} href={url}>
                    <Button
                      icon={<ExportOutlined style={{ fontSize: 20 }} />}
                      size={'small'}
                      shape={'circle'}
                      style={{ border: 'none', background: 'none' }}
                    />
                  </a>
                ) : null,
              props: {
                title: ''
              }
            };
          }
        }
      },
      {
        title: 'Control',
        key: 'Control',
        width: '5em',
        align: 'center',
        render: (text, record) => {
          let lab = props.instances.find(
            lab => lab.metadata.name === record.key
          );
          let template = props.templates.find(
            item =>
              item.metadata.name ===
              lab?.spec['template.crownlabs.polito.it/TemplateRef'].name
          );
          return {
            children:
              template &&
              template.spec.environmentList[0].persistent ?
              ((lab?.status?.phase === 'VmiReady' ||
                lab?.status?.phase === 'VmiOff') &&
              (lab?.spec.running ? (
                <Popconfirm
                  title={'Shutdown VM?'}
                  onConfirm={() => {
                    const shutDownLab = { ...lab };
                    shutDownLab.spec.running = false;
                    window.api.updateGenericResource(
                      lab.metadata.selfLink,
                      shutDownLab
                    );
                  }}
                >
                  <Button
                    icon={
                      <PauseCircleOutlined
                        style={{ fontSize: 20, color: colorYellow }}
                      />
                    }
                    size={'small'}
                    shape={'circle'}
                    style={{ border: 'none', background: 'none' }}
                  />
                </Popconfirm>
              ) : (
                <Tooltip title={'Start VM?'}>
                  <Button
                    icon={
                      <PlayCircleOutlined
                        style={{ fontSize: 20, color: colorBlue }}
                      />
                    }
                    onClick={() => {
                      const startedLab = { ...lab };
                      startedLab.spec.running = true;
                      window.api.updateGenericResource(
                        lab.metadata.selfLink,
                        startedLab
                      );
                    }}
                    size={'small'}
                    shape={'circle'}
                    style={{ border: 'none', background: 'none' }}
                  />
                </Tooltip>
              ))) : (
                template && template.spec.environmentList[0].environmentType === 'Container'
                  && lab.status && lab.status.url && (
                <Tooltip title={'Open file browser'}>
                  <a target={'_blank'} href={`${lab.status.url}/mydrive/files`}>
                    <Button
                      icon={<FolderOpenOutlined style={{ fontSize: 20 }} />}
                      size={'small'}
                      shape={'circle'}
                      style={{ border: 'none', background: 'none' }}
                    />
                  </a>
                </Tooltip>
              )),
            props: {
              title: ''
            }
          };
        }
      },
      {
        title: 'Destroy',
        key: 'Destroy',
        width: '5em',
        align: 'center',
        render: (text, record) => {
          let lab = props.instances.find(
            lab => lab.metadata.name === record.key
          );
          return {
            children: (
              <Popconfirm
                title={'Destroy VM?'}
                onConfirm={() => {
                  setTerminating(prev => {
                    prev.push(record.key);
                    return [...prev];
                  });
                  window.api.deleteGenericResource(lab.metadata.selfLink);
                }}
              >
                <Button
                  icon={
                    <DeleteOutlined style={{ fontSize: 20, color: colorRed }} />
                  }
                  size={'small'}
                  shape={'circle'}
                  style={{ border: 'none', background: 'none' }}
                />
              </Popconfirm>
            ),
            props: {
              title: ''
            }
          };
        }
      },
      {
        dataIndex: 'Age',
        key: 'Age',
        title: 'Age',
        width: '5em',
        sorter: {
          compare: (a, b) => compareAge(a.Age, b.Age)
        },
        last: true,
        render: text => {
          return {
            children: text,
            props: {
              title: ''
            }
          };
        }
      }
    ]);
  }, [props.instances, props.templates, terminating]);

  const renderInstances = (text, record, dataIndex) => {
    return dataIndex === 'Name' ? (
      <Typography.Text strong>{text}</Typography.Text>
    ) : (
      <div>{text}</div>
    );
  };

  const instancesViews = [];
  props.instances.forEach(instances => {
    let template = props.templates.find(
      item =>
        item.metadata.name ===
        instances.spec['template.crownlabs.polito.it/TemplateRef'].name
    );
    if (template) {
      instancesViews.push({
        key: instances.metadata.name,
        Status: instances.status && instances.status.phase,
        Name:
          template &&
          template.spec.description +
            ' - ' +
            instances.metadata.name.split('-').slice(-1),
        User: props.onProfessor
          ? instances.spec['tenant.crownlabs.polito.it/TenantRef'].name
          : null,
        IP: instances.status ? instances.status.ip : '',
        Age: calculateAge(instances.metadata.creationTimestamp),
        'VM Type': template.spec.environmentList[0].guiEnabled,
        IsPersistent: template.spec.environmentList[0].persistent
      });
    }
  });

  return (
    <>
      <Table
        columns={columns}
        dataSource={instancesViews}
        pagination={{
          position: ['bottomCenter'],
          hideOnSinglePage: instancesViews.length < 11,
          showSizeChanger: true
        }}
        showSorterTooltip={false}
        components={{
          header: {
            cell: ResizableTitle
          }
        }}
        scroll={{ x: 'max-content' }}
        loading={props.loading}
        locale={{
          emptyText: (
            <Empty
              description={'No Running Labs'}
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )
        }}
      />
    </>
  );
}
