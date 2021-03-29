import React, { useEffect, useState } from 'react';
import { withRouter, useLocation } from 'react-router-dom';
import { notification, Alert, Badge, Button, Tooltip, message } from 'antd';
import DraggableLayout from '../../widgets/draggableLayout/DraggableLayout';
import Templates, { colorBlue } from './Templates';
import Instances from './Instances';
import Utils from '../../services/Utils';
import { PlusOutlined } from '@ant-design/icons';
import TemplateForm from './TemplateForm';

function CrownLabsHome(props) {
  const [loading, setLoading] = useState(true);
  const [openCreate, setOpenCreate] = useState(false);
  const [templates, setTemplates] = useState([]);
  const [instances, setInstances] = useState([]);
  const [tenants, setTenants] = useState([]);
  const [instancesProfessor, setInstancesProfessor] = useState([]);
  let location = useLocation();
  const onProfessor = location.pathname === '/professor';

  useEffect(() => {
    if (onProfessor && instances.length > 0) {
      let workspaces = [];
      instances.forEach(instance => {
        templates.forEach(template => {
          if (
            instance.spec['template.crownlabs.polito.it/TemplateRef'].name ===
            template.metadata.name
          ) {
            workspaces.push(instance);
          }
        });
      });
      setInstancesProfessor(workspaces);
    }
  }, [instances]);

  useEffect(() => {
    if (tenants.length === 1) {
      if (onProfessor) {
        window.api.setNamespace('all namespaces');

        let managedWorkspaces = tenants[0].spec.workspaces.filter(
          workspace => workspace.role === 'manager'
        );

        if (managedWorkspaces.length > 0) {
          managedWorkspaces.forEach(workspace => {
            setLoading(true);
            window.api
              .getGenericResource(
                '/apis/crownlabs.polito.it/v1alpha2/namespaces/workspace-' +
                  workspace.workspaceRef.name +
                  '/templates',
                setTemplates,
                true
              )
              .then(res => {
                setTemplates(prev => [...prev, ...res.items]);
                setLoading(false);
                return res.items;
              })
              .then(() => {
                window.api
                  .getGenericResource(
                    '/apis/crownlabs.polito.it/v1alpha2/instances',
                    setInstances
                  )
                  .then(res => {
                    setInstances(res.items);
                  })
                  .catch(() => setLoading(false));
              })
              .catch(() => setLoading(false));
          });
        } else {
          setLoading(false);
        }
      } else {
        window.api.setNamespace(tenants[0].status.personalNamespace.name);

        tenants[0].spec.workspaces.forEach(workspace => {
          setLoading(true);
          window.api
            .getGenericResource(
              '/apis/crownlabs.polito.it/v1alpha2/namespaces/workspace-' +
                workspace.workspaceRef.name +
                '/templates',
              setTemplates,
              true
            )
            .then(res => {
              setTemplates(prev => [...prev, ...res.items]);
              setLoading(false);
            })
            .catch(() => setLoading(false));
        });
        window.api
          .getGenericResource(
            '/apis/crownlabs.polito.it/v1alpha2/namespaces/' +
              tenants[0].status.personalNamespace.name +
              '/instances',
            setInstances
          )
          .then(res => {
            setInstances(res.items);
          })
          .catch(() => setLoading(false));
      }
    }
  }, [tenants]);

  useEffect(() => {
    if (Utils().parseJWT()) {
      window.api
        .getGenericResource(
          '/apis/crownlabs.polito.it/v1alpha1/tenants/' +
            Utils().parseJWT().preferred_username,
          setTenants,
          true
        )
        .then(res => setTenants([res]))
        .catch(() => {
          setLoading(false);
          message.error(
            'Failed to get tenant ' + Utils().parseJWT().preferred_username
          );
        });
    } else {
      setLoading(false);
      message.error('Impossible to parse token');
    }

    /**
     * Delete any reference to the component in the api service.
     * Avoid no-op and memory leaks
     */
    return () => {
      window.api.abortWatch('instances');
      window.api.abortWatch('templates');
      window.api.abortWatch('tenants');
    };
  }, []);

  useEffect(() => {
    const FORM_CHANNEL_UPD = 'FORM_CHANNEL_UPD';
    const localUpd = localStorage.getItem(FORM_CHANNEL_UPD);

    if (!localUpd) {
      notification.open({
        message: '',
        description: (
          <div
            style={{
              textAlign: 'center',
              margin: 20
            }}
          >
            We would like to know your thoughts about CrownLabs! Please fill{' '}
            <a
              target="_blank"
              href="https://forms.gle/g86xkWTHoaULfPHp8"
              style={{}}
            >
              this form
            </a>
            !
            <br /> If you would like to keep posted about the latest news of
            CrownLabs, join our{' '}
            <a target="_blank" href="https://t.me/crownlabsNews">
              channel
              <img
                style={{ display: 'inline', height: 25, margin: 5 }}
                src="https://cdn.svgporn.com/logos/telegram.svg"
              />
            </a>
          </div>
        ),
        duration: 0,
        placement: 'topRight',
        onClose: () => {
          localStorage.setItem(FORM_CHANNEL_UPD, JSON.stringify(true));
        },
        style: {
          border: `4px solid ${colorBlue}`,
          padding: 10
        }
      });
    }
  }, []);

  const items = [];

  items.push(
    <div
      data-grid={{
        lg: { w: 10, h: 38, x: 0, y: 0, minH: 28 },
        md: { w: 24, h: 38, x: 0, y: 0, minH: 28 }
      }}
      key={'table_templates'}
      title={<Badge text={'Available Images'} color={'blue'} />}
      extra={
        onProfessor
          ? [
              <Tooltip title={'Create new template'} key={'add_template'}>
                <Button
                  icon={<PlusOutlined />}
                  style={{ marginTop: -8, marginBottom: -8, marginRight: -8 }}
                  type={'primary'}
                  onClick={() => setOpenCreate(true)}
                />
              </Tooltip>
            ]
          : null
      }
    >
      <Templates
        loading={loading}
        {...props}
        tenants={tenants}
        onProfessor={onProfessor}
        templates={templates}
      />
    </div>,
    <div
      data-grid={{
        lg: { w: 14, h: 38, x: 10, y: 0, minH: 28 },
        md: { w: 24, h: 38, x: 10, y: 0, minH: 28 }
      }}
      key={'table_labinstances'}
      title={<Badge text={'Running Images'} color={'blue'} />}
    >
      <Instances
        loading={loading}
        {...props}
        onProfessor={onProfessor}
        templates={templates}
        instances={onProfessor ? instancesProfessor : instances}
      />
    </div>
  );

  return (
    <Alert.ErrorBoundary>
      <DraggableLayout
        title={<Badge text={'templates'} color={'blue'} />}
        breakpoints={{
          lg: 1300,
          md: 796,
          sm: 568,
          xs: 280,
          xss: 0
        }}
        responsive
      >
        {items}
      </DraggableLayout>
      <TemplateForm
        templates={templates}
        visible={openCreate}
        setVisible={setOpenCreate}
      />
      <div
        style={{
          textAlign: 'center',
          margin: 'auto',
          marginTop: 20,
          padding: '0 20px',
          maxWidth: 800
        }}
      >
        <h1
          style={{
            fontSize: '1.2rem'
          }}
        >
          Persistent VMs are here!
        </h1>{' '}
        <p
          style={{
            fontSize: '1rem'
          }}
        >
          You can now shutdown persistent VMs without destroying them, to avoid
          resource consumption. Once off you can start them back up. Creating
          persistent VMs could take more time, around 6-8 minutes.
        </p>
      </div>
    </Alert.ErrorBoundary>
  );
}

export default withRouter(CrownLabsHome);
