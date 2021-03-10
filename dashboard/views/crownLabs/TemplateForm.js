import { withRouter } from 'react-router-dom';
import React, { useEffect, useRef, useState } from 'react';
import { Button, Form, Input, message, Modal, Select, Slider } from 'antd';
import Utils from '../../services/Utils';
import { CodeOutlined, DesktopOutlined } from '@ant-design/icons';

function TemplateForm(props) {
  const [courses, setCourses] = useState([]);
  const [selectedCourse, setSelectedCourse] = useState(null);
  const [images, setImages] = useState([]);
  const [selectedImage, setSelectedImage] = useState(null);
  const [versions, setVersions] = useState([]);
  const [selectedVersion, setSelectedVersion] = useState(null);
  const [cpu, setCpu] = useState(2);
  const [cpuPerc, setCpuPerc] = useState(33);
  const [memory, setMemory] = useState(2);

  const description = useRef('');
  const VMType = useRef('');
  const EnvType = useRef('');
  const registry = useRef('');

  const addTemplate = () => {
    const namespace = 'workspace-' + selectedCourse;
    const labNums = props.templates.filter(
      item =>
        item.metadata.namespace === namespace &&
        item.metadata.name.includes(selectedCourse + '-lab')
    ).length;
    let labNumber = labNums + 1;

    if (
      !selectedCourse ||
      !selectedImage ||
      !selectedVersion ||
      description.current === '' ||
      VMType.current === ''
    ) {
      message.error('Please fill all the fields');
      return;
    }

    const item = {
      apiVersion: 'crownlabs.polito.it/v1alpha2',
      kind: 'Template',
      metadata: {
        name: selectedCourse + '-lab' + labNumber,
        namespace
      },
      spec: {
        description: description.current,
        prettyName: description.current,
        ['workspace.crownlabs.polito.it/WorkspaceRef']: {
          name: selectedCourse
        },
        environmentList: [
          {
            environmentType: EnvType.current,
            guiEnabled: VMType.current,
            image:
              registry.current + '/' + selectedImage + ':' + selectedVersion,
            name: selectedCourse + '-lab' + labNumber,
            persistent: false,
            resources: {
              cpu: cpu,
              memory: memory + 'G',
              reservedCPUPercentage: cpuPerc
            }
          }
        ]
      }
    };

    window.api
      .createCustomResource(
        'crownlabs.polito.it',
        'v1alpha2',
        namespace,
        'templates',
        item
      )
      .then(() => props.setVisible(false));
  };

  useEffect(() => {
    setVersions([]);
    setSelectedVersion([]);
    if (selectedImage) {
      images
        .find(item => item.value === selectedImage)
        .versions.forEach(v => setVersions(prev => [...prev, { value: v }]));

      if (
        images.find(item => item.value === selectedImage).versions.length === 1
      )
        setSelectedVersion(
          images.find(item => item.value === selectedImage).versions[0]
        );
    }
  }, [selectedImage]);

  useEffect(() => {
    window.api
      .getGenericResource(
        '/apis/crownlabs.polito.it/v1alpha1/imagelists/crownlabs-virtual-machine-images'
      )
      .then(res => {
        registry.current = res.spec.registryName;
        res.spec.images.forEach(image => {
          setImages(prev => [
            ...prev,
            { value: image.name, versions: image.versions }
          ]);
        });
      })
      .catch(() => {});

    if (props.templates.length !== 0) {
      let workspaces = [];
      props.templates.forEach(template => {
        if (
          !workspaces.includes(
            template.spec['workspace.crownlabs.polito.it/WorkspaceRef'].name
          )
        )
          workspaces.push(
            template.spec['workspace.crownlabs.polito.it/WorkspaceRef'].name
          );
      });

      setCourses([]);

      workspaces.forEach(workspace => {
        setCourses(prev => [
          ...prev,
          {
            value: workspace
          }
        ]);
      });

      if (workspaces.length === 1) setSelectedCourse(workspaces[0]);
    }
  }, [props.templates]);

  const formItemLayout = {
    labelCol: { span: 6 },
    wrapperCol: { span: 18 }
  };

  return (
    <Modal
      title={'Create new template'}
      visible={props.visible}
      onCancel={() => props.setVisible(false)}
      onOk={addTemplate}
    >
      <Form {...formItemLayout}>
        <Form.Item label={'Course'}>
          <Select
            options={courses}
            value={selectedCourse}
            placeholder={'Code'}
            showSearch
            onSelect={value => setSelectedCourse(value)}
          />
        </Form.Item>
        <Form.Item label={'Image'}>
          <Input.Group compact>
            <Select
              options={images}
              value={selectedImage}
              style={{ width: '70%' }}
              showSearch
              placeholder={'Name'}
              onSelect={value => setSelectedImage(value)}
            />
            <Select
              options={versions}
              value={selectedVersion}
              style={{ width: '30%' }}
              showSearch
              placeholder={'Version'}
              onSelect={value => setSelectedVersion(value)}
            />
          </Input.Group>
        </Form.Item>
        <Form.Item label={'Env Type'}>
          <Select
            placeholder={'Environment Type'}
            onSelect={value => (EnvType.current = value)}
          >
            <Select.Option value={'Container'}>Container</Select.Option>
            <Select.Option value={'VirtualMachine'}>
              Virtual Machine
            </Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label={'VM'}>
          <Input.Group compact>
            <Input
              placeholder={'Description'}
              style={{ width: '70%' }}
              onChange={e => (description.current = e.target.value)}
            />
            <Select
              placeholder={'Type'}
              style={{ width: '30%' }}
              onSelect={value => (VMType.current = value)}
            >
              <Select.Option value={false}>
                <CodeOutlined /> CLI
              </Select.Option>
              <Select.Option value={true}>
                <DesktopOutlined /> GUI
              </Select.Option>
            </Select>
          </Input.Group>
        </Form.Item>
        <Form.Item label={'CPU (Cores)'}>
          <Slider
            min={1}
            max={4}
            defaultValue={cpu}
            onChange={value => setCpu(value)}
            dots
          />
        </Form.Item>
        <Form.Item label={'Reserved CPU (%)'}>
          <Slider
            min={0}
            max={100}
            defaultValue={cpuPerc}
            onChange={value => setCpuPerc(value)}
          />
        </Form.Item>
        <Form.Item label={'Memory (GB)'}>
          <Slider
            min={0.5}
            max={8}
            defaultValue={memory}
            step={0.5}
            onChange={value => setMemory(value)}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default withRouter(TemplateForm);
