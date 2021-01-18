import React, { useEffect, useState } from 'react';
import { withRouter, useLocation } from 'react-router-dom';
import { Row, Alert, Badge, Button, Tooltip } from 'antd';
import DraggableLayout from '../../widgets/draggableLayout/DraggableLayout';
import Templates from './Templates';
import Instances from './Instances';
import Utils from '../../services/Utils';
import { PlusOutlined } from '@ant-design/icons';
import TemplateForm from './TemplateForm';

function CrownLabsHome(props){

  const [loading, setLoading] = useState(true);
  const [openCreate, setOpenCreate] = useState(false);
  const [templates, setTemplates] = useState([]);
  const [instances, setInstances] = useState([]);
  const [tenants, setTenants] = useState([]);
  const [instancesProfessor, setInstancesProfessor] = useState([]);
  let location = useLocation();
  const onProfessor = location.pathname === '/professor';

  useEffect(() => {
    if(onProfessor && instances.length > 0){
      let workspaces = [];
      instances.forEach(instance => {
        templates.forEach(template => {
          if(instance.spec['template.crownlabs.polito.it/TemplateRef'].name === template.metadata.name){
            workspaces.push(instance);
          }
        })
      })
      setInstancesProfessor(workspaces);
    }
  }, [instances])

  useEffect(() => {
    if(tenants.length === 1){
      if(onProfessor){
        window.api.setNamespace('all namespaces')

        tenants[0].spec.workspaces.filter(workspace => workspace.role === 'manager')
          .forEach(workspace => {
            setLoading(true);
            window.api.getGenericResource('/apis/crownlabs.polito.it/v1alpha2/namespaces/workspace-' + workspace.workspaceRef.name + '/templates', setTemplates, true)
              .then(res =>{
                setTemplates(prev => [...prev, ...res.items])
                setLoading(false);
                return res.items;
              }).then(() => {
                window.api.getGenericResource('/apis/crownlabs.polito.it/v1alpha2/instances', setInstances)
                  .then(res =>{
                    setInstances(res.items);
                  }).catch(() => setLoading(false))
            }).catch(() => setLoading(false))
          });
      } else {
        window.api.setNamespace(tenants[0].status.personalNamespace.name);

        tenants[0].spec.workspaces.forEach(workspace => {
          setLoading(true);
          window.api.getGenericResource('/apis/crownlabs.polito.it/v1alpha2/namespaces/workspace-' + workspace.workspaceRef.name + '/templates', setTemplates, true)
            .then(res =>{
              setTemplates(prev => [...prev, ...res.items])
              setLoading(false);
            }).catch(() => setLoading(false))
        });
        window.api.getGenericResource('/apis/crownlabs.polito.it/v1alpha2/namespaces/' + tenants[0].status.personalNamespace.name + '/instances', setInstances)
          .then(res =>{
            setInstances(res.items);
          }).catch(() => setLoading(false))
      }
    }
  }, [tenants])

  useEffect(() => {
    if(Utils().parseJWT()){
      window.api.getGenericResource('/apis/crownlabs.polito.it/v1alpha1/tenants/' + Utils().parseJWT().preferred_username, setTenants, true)
        .catch(() => setLoading(false))
    }

    /**
     * Delete any reference to the component in the api service.
     * Avoid no-op and memory leaks
     */
    return () => {
      window.api.abortWatch('instances');
      window.api.abortWatch('templates');
      window.api.abortWatch('tenants');
    }
  }, []);

  const items = [];

  items.push(
    <div data-grid={{ w: 10, h: 38, x: 0, y: 0 }}
         key={'table_templates'}
         title={<Badge text={'Available Images'} color={'blue'} />}
         extra={onProfessor ? [
           <Tooltip title={'Create new template'} key={'add_template'}>
             <Button icon={<PlusOutlined />}
                     style={{marginTop: -8, marginBottom: -8, marginRight: -8}}
                     type={'primary'} onClick={() => setOpenCreate(true)} />
           </Tooltip>
         ] : null}
    >
      <Templates loading={loading}
                    {...props}
                    tenants={tenants}
                    onProfessor={onProfessor}
                    templates={templates}
      />
    </div>,
    <div data-grid={{ w: 14, h: 38, x: 10, y: 0 }}
         key={'table_instances'}
         title={<Badge text={'Running Images'} color={'blue'} />}
    >
      <Instances loading={loading}
                 {...props}
                 onProfessor={onProfessor}
                 templates={templates}
                 instances={onProfessor ? instancesProfessor : instances}
      />
    </div>
  )

  return(
    <Alert.ErrorBoundary>
      <DraggableLayout title={
        <Badge text={'templates'} color={'blue'} />
      }
      >
        {items}
      </DraggableLayout>
      <TemplateForm templates={templates}
                    visible={openCreate}
                    setVisible={setOpenCreate}
      />
    </Alert.ErrorBoundary>
  );
}

export default withRouter(CrownLabsHome);
