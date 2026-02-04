import { type FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';

import { useContext, useEffect, useMemo, useState } from 'react';
import { type FC } from 'react';
import {
  type UpdatedWorkspaceTemplatesSubscription,
  UpdateType,
  useCreateInstanceMutation,
  useDeleteTemplateMutation,
  useWorkspaceTemplatesQuery,
  type UpdatedWorkspaceTemplatesSubscriptionResult,
  EnvironmentType,
  type EnvironmentListListItemInput,
  type SharedVolumeMountsListItemInput,
  useApplyTemplateJsonPatchMutation,
} from '../../../../generated-types';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import { updatedWorkspaceTemplates } from '../../../../graphql-components/subscription';
import { type Template, WorkspaceRole } from '../../../../utils';
import { ErrorTypes } from '../../../../errorHandling/utils';
import {
  makeGuiTemplate,
  joinInstancesAndTemplates,
} from '../../../../utilsLogic';
import { TemplatesEmpty } from '../TemplatesEmpty';
import { TemplatesTable } from '../TemplatesTable';
import { SharedVolumesDrawer } from '../../SharedVolumes';
import { AuthContext } from '../../../../contexts/AuthContext';
import ModalCreateTemplate from '../../ModalCreateTemplate';
import type { TemplateForm } from '../../ModalCreateTemplate/types';
import { getImageNameNoVer } from '../../ModalCreateTemplate/utils';
import { OwnedInstancesContext } from '../../../../contexts/OwnedInstancesContext';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  workspaceName: string;
  role: WorkspaceRole;
  isPersonal?: boolean;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';
const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);
  const {
    tenantNamespace,
    workspaceNamespace,
    workspaceName,
    role,
    isPersonal,
  } = props;

  // Get instances from context
  const { instances: ownedInstances } = useContext(OwnedInstancesContext);

  const {
    loading: loadingTemplate,
    error: errorTemplate,
    subscribeToMore: subscribeToMoreTemplates,
    data: templateListData,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    onError: error => {
      console.error(
        'TemplatesTableLogic useWorkspaceTemplatesQuery error:',
        error,
        'workspaceNamespace:',
        workspaceNamespace,
      );
      apolloErrorCatcher(error);
    },
    fetchPolicy: fetchPolicy_networkOnly,
    nextFetchPolicy: 'cache-only',
  });

  const dataTemplate = useMemo(() => {
    const templates =
      templateListData?.templateList?.templates
        ?.map(t =>
          makeGuiTemplate({
            original: t ?? {},
            alias: {
              id: t?.metadata?.name ?? '',
              name: t?.spec?.prettyName ?? '',
            },
          }),
        )
        .sort((a, b) => a.name.localeCompare(b.name)) ?? [];
    return templates;
  }, [templateListData?.templateList?.templates]);

  useEffect(() => {
    if (!loadingTemplate && !errorTemplate && !errorsQueue.length) {
      const unsubscribe =
        subscribeToMoreTemplates<UpdatedWorkspaceTemplatesSubscription>({
          onError: makeErrorCatcher(ErrorTypes.GenericError),
          document: updatedWorkspaceTemplates,
          variables: { workspaceNamespace },
          updateQuery: (prev, { subscriptionData }) => {
            const { data } = subscriptionData;
            if (!data?.updatedTemplate) return prev;

            const { template, updateType } = data.updatedTemplate;
            const templates = prev.templateList?.templates ?? [];

            let out = [] as NonNullable<
              NonNullable<
                UpdatedWorkspaceTemplatesSubscriptionResult['data']
              >['updatedTemplate']
            >['template'][];

            switch (updateType) {
              case UpdateType.Added:
                // Only process if template data is valid
                if (template) {
                  out = [...templates, template];
                } else {
                  out = templates;
                }
                break;
              case UpdateType.Modified:
                // Only process if template data is valid
                if (template) {
                  out = templates.map(t =>
                    t?.metadata?.name === template.metadata?.name
                      ? template
                      : t,
                  );
                } else {
                  out = templates;
                }
                break;
              case UpdateType.Deleted:
                // For deletions, we only need the template metadata (name) to filter
                // Don't try to access template.spec or other potentially malformed data
                if (template?.metadata?.name) {
                  out = templates.filter(
                    t => t?.metadata?.name !== template.metadata?.name,
                  );
                } else {
                  out = templates;
                }
                break;
              default:
                out = templates;
                break;
            }

            const result = Object.assign({}, prev, {
              templateList: {
                templates: out,
                __typename: prev.templateList?.__typename,
              },
            });
            return result;
          },
        });
      return unsubscribe;
    }
  }, [
    errorTemplate,
    errorsQueue.length,
    loadingTemplate,
    subscribeToMoreTemplates,
    userId,
    workspaceNamespace,
    apolloErrorCatcher,
    makeErrorCatcher,
  ]);

  const [createInstanceMutation] = useCreateInstanceMutation({
    onError: apolloErrorCatcher,
  });
  const [deleteTemplateMutation, { loading: loadingDeleteTemplateMutation }] =
    useDeleteTemplateMutation({
      onError: apolloErrorCatcher,
    });

  const createInstance = (templateId: string, labelSelector?: JSON) =>
    createInstanceMutation({
      variables: {
        templateId,
        tenantNamespace,
        tenantId: userId ?? '',
        workspaceNamespace,
        nodeSelector: labelSelector as Record<string, string> | undefined,
      },
    }).catch(error => {
      console.error('TemplatesTableLogic createInstance error:', error);
      throw error;
    });

  const templates = useMemo(() => {
    const joined = joinInstancesAndTemplates(dataTemplate, ownedInstances);

    // build map of original GraphQL templates by metadata.name for reliable lookup
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const originalById = new Map<string, any>();
    (templateListData?.templateList?.templates ?? []).forEach(t => {
      const id = t?.metadata?.name;
      if (id) originalById.set(id, t);
    });

    // Enrich joined templates using the original template spec when available
    return (joined || []).map(t => {
      const id = t?.id;
      const original = id ? originalById.get(id) : undefined;
      const env = original?.spec?.environmentList?.[0];
      return {
        ...t,
        image: env?.image ?? null,
        environmentType: env?.environmentType ?? null,
      };
    });
  }, [dataTemplate, ownedInstances, templateListData?.templateList?.templates]);

  const [showTemplateModal, setShowTemplateModal] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<TemplateForm>();

  const [applyTemplateJsonPatchMutation] = useApplyTemplateJsonPatchMutation({
    onError: apolloErrorCatcher,
  });

  const [usedTemplate, setUsedTemplate] = useState<Template | null>(null);

  const submitPatchHandler = async (t: TemplateForm) => {
    try {
      // const patchJson = getTemplatePatchJson({
      //   spec: {
      //     prettyName: t.name,
      //     deleteAfter: t.deleteAfter,
      //     inactivityTimeout: t.inactivityTimeout,
      //     description: usedTemplate?.description ?? t.name,
      //     environmentList: t.environments.map(
      //       (env): EnvironmentListListItemInput => ({
      //         name: env.name,
      //         mountMyDriveVolume: usedTemplate?.environmentList.find(e => e.name === env.name)?.mountMyDriveVolume ?? true,
      //         guiEnabled: env.gui,
      //         persistent: env.persistent,
      //         environmentType: env.environmentType,
      //         resources: {
      //           reservedCPUPercentage: usedTemplate?.environmentList.find(e => e.name === env.name)?.resources.reservedCPUPercentage ?? 50,
      //           cpu: env.cpu,
      //           memory: `${env.ram * 1000}Mi`, // convert Gi to Mi
      //           disk: env.disk ? `${env.disk * 1000}Mi` : undefined, // convert Gi to Mi
      //         },
      //         image: env.registry
      //           ? `${env.registry}/${env.image}`
      //           : env.image,
      //         sharedVolumeMounts: (env.sharedVolumeMounts ?? []).map(
      //           (svm): SharedVolumeMountsListItemInput => ({
      //             mountPath: svm.mountPath,
      //             readOnly: svm.readOnly,
      //             sharedVolume: {
      //               name: svm.sharedVolume,
      //               namespace: workspaceNamespace,
      //             }
      //           }),
      //         ),
      //       }),
      //     ),
      //   },
      // });

      const environmentList = t.environments.map(
        (env): EnvironmentListListItemInput => ({
          name: env.name,
          mountMyDriveVolume:
            usedTemplate?.environmentList.find(e => e.name === env.name)
              ?.mountMyDriveVolume ?? true,
          guiEnabled: env.gui,
          persistent: env.persistent,
          environmentType: env.environmentType,
          resources: {
            reservedCPUPercentage:
              usedTemplate?.environmentList.find(e => e.name === env.name)
                ?.resources.reservedCPUPercentage ?? 50,
            cpu: env.cpu,
            memory: `${env.ram * 1000}Mi`, // convert Gi to Mi
            disk: env.disk ? `${env.disk * 1000}Mi` : undefined, // convert Gi to Mi
          },
          image: env.registry ? `${env.registry}/${env.image}` : env.image,
          sharedVolumeMounts: (env.sharedVolumeMounts ?? []).map(
            (svm): SharedVolumeMountsListItemInput => ({
              mountPath: svm.mountPath,
              readOnly: svm.readOnly,
              sharedVolume: {
                name: svm.sharedVolume,
                namespace: workspaceNamespace,
              },
            }),
          ),
        }),
      );

      const patchJson = JSON.stringify([
        { op: 'replace', path: '/spec/environmentList', value: environmentList },
      { op: 'replace', path: '/spec/prettyName', value: t.name },
      { op: 'replace', path: '/spec/deleteAfter', value: t.deleteAfter },
      { op: 'replace', path: '/spec/inactivityTimeout', value: t.inactivityTimeout },
      { op: 'replace', path: '/spec/allowPublicExposure', value: t.allowPublicExposure },
      { op: 'replace', path: '/spec/description', value: t.description },
      ]);


      //    console.log('Patch JSON:', patchJson);

      return await applyTemplateJsonPatchMutation({
        variables: {
          workspaceNamespace,
          templateId: usedTemplate?.id ?? '',
          patchJson: patchJson,
          manager: 'frontend-template-patch',
        },
      });
    } catch (error) {
      console.error('TemplatesTableLogic applyTemplateMutation error:', error);
      throw error;
    }
  };

  return (
    <div
      style={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        flex: '1 1 auto',
        minHeight: 0,
      }}
    >
      {/* full-height flex column so TemplatesTable can take the remaining space and scroll */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          flex: '1 1 auto',
          minHeight: 0,
        }}
      >
        <Spin
          size="large"
          spinning={loadingTemplate}
          style={{
            display: 'flex',
            flexDirection: 'column',
            flex: '1 1 auto',
            minHeight: 0,
          }}
        >
          {!loadingTemplate && !errorTemplate && templates && ownedInstances ? (
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                flex: '1 1 auto',
                minHeight: 0,
              }}
            >
              <TemplatesTable
                tenantNamespace={tenantNamespace}
                workspaceNamespace={workspaceNamespace}
                workspaceName={workspaceName}
                templates={templates}
                role={role}
                deleteTemplate={(templateId: string) =>
                  deleteTemplateMutation({
                    variables: {
                      workspaceNamespace,
                      templateId,
                    },
                  })
                }
                deleteTemplateLoading={loadingDeleteTemplateMutation}
                editTemplate={(template: Template) => {
                  setUsedTemplate(template);
                  const templateForm: TemplateForm = {
                    name: template.name,
                    nodeSelector: template.nodeSelector ?? {},
                    description: template.description ?? template.name,
                    deleteAfter: template.deleteAfter,
                    allowPublicExposure: template.allowPublicExposure,
                    inactivityTimeout: template.inactivityTimeout,
                    environments: template.environmentList.map(env => ({
                      name: env.name,
                      persistent: env.persistent,
                      environmentType:
                        env.environmentType ?? EnvironmentType.VirtualMachine,
                      cpu: env.resources.cpu,
                      ram: parseInt(env.resources.memory) / 1000, // assuming memory is in 'XMi' format
                      disk: env.resources.disk
                        ? parseInt(env.resources.disk) / 1000
                        : 0, // convert from Mi to Gi
                      image:
                        getImageNameNoVer(env.image)
                          .split('/')
                          .slice(-2)
                          .join('/') ?? '',
                      registry:
                        getImageNameNoVer(env.image).split('/').slice(0)[0] ??
                        '',
                      sharedVolumeMounts: env.sharedVolumeMounts.map(svm => ({
                        sharedVolume: svm.name,
                        mountPath: svm.mountPath,
                        readOnly: svm.readOnly,
                      })),
                      rewriteUrl: false,
                      gui: env.guiEnabled,
                    })),
                  };
                  setEditingTemplate(templateForm);
                  setShowTemplateModal(true);
                }}
                createInstance={createInstance}
                isPersonal={isPersonal}
              />
              <ModalCreateTemplate
                show={showTemplateModal}
                setShow={setShowTemplateModal}
                template={editingTemplate}
                workspaceNamespace={workspaceNamespace}
                cpuInterval={{ max: 8, min: 1 }}
                ramInterval={{ max: 32, min: 1 }}
                diskInterval={{ max: 50, min: 10 }}
                submitHandler={submitPatchHandler}
                loading={false}
                isPersonal={isPersonal}
              />
            </div>
          ) : (
            <div
              className={
                loadingTemplate || errorTemplate ? 'invisible' : 'visible'
              }
              style={{
                flex: '1 1 auto',
                minHeight: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <TemplatesEmpty role={role} />
            </div>
          )}
        </Spin>
      </div>
      {role === WorkspaceRole.manager && !loadingTemplate && !isPersonal ? (
        <div
          style={{
            position: 'sticky',
            bottom: 0,
            zIndex: 100,
          }}
          className="cl-shared-volumes-bg"
        >
          <SharedVolumesDrawer
            workspaceNamespace={workspaceNamespace}
            isPersonal={isPersonal}
          />
        </div>
      ) : null}
    </div>
  );
};

export default TemplatesTableLogic;
