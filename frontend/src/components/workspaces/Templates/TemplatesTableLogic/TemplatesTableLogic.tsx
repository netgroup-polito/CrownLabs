import { type FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';

import { useContext, useEffect, useMemo, useState } from 'react';
import { type FC } from 'react';
import {
  type UpdatedWorkspaceTemplatesSubscription,
  UpdateType,
  useCreateInstanceMutation,
  useDeleteTemplateMutation,
  useOwnedInstancesQuery,
  useWorkspaceTemplatesQuery,
  type UpdatedWorkspaceTemplatesSubscriptionResult,
} from '../../../../generated-types';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import {
  updatedOwnedInstances,
  updatedWorkspaceTemplates,
} from '../../../../graphql-components/subscription';
import { type Instance, WorkspaceRole } from '../../../../utils';
import { ErrorTypes } from '../../../../errorHandling/utils';
import {
  makeGuiInstance,
  makeGuiTemplate,
  joinInstancesAndTemplates,
  updateQueryOwnedInstancesQuery,
} from '../../../../utilsLogic';
import { TemplatesEmpty } from '../TemplatesEmpty';
import { TemplatesTable } from '../TemplatesTable';
import { SharedVolumesDrawer } from '../../SharedVolumes';
import { AuthContext } from '../../../../contexts/AuthContext';
import { TenantContext } from '../../../../contexts/TenantContext';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  workspaceName: string;
  role: WorkspaceRole;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
  refreshQuota?: () => void; // Add refresh function
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
    availableQuota,
    refreshQuota,
    isPersonal,
  } = props;

  const [dataInstances, setDataInstances] = useState<Instance[]>([]);

  // Add the missing instances query
  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
    onError: apolloErrorCatcher,
    onCompleted: data => {
      const instances =
        data?.instanceList?.instances
          ?.map(i => {
            const guiInstance = makeGuiInstance(i, userId);
            return guiInstance;
          })
          .filter(Boolean) ?? [];
      setDataInstances(instances);
    },
    fetchPolicy: fetchPolicy_networkOnly,
    nextFetchPolicy: 'cache-only',
  });

  // Subscribe to instance updates
  const notifier = useContext(TenantContext).notify;

  useEffect(() => {
    if (!loadingInstances && !errorInstances && !errorsQueue.length) {
      const unsubscribe = subscribeToMoreInstances({
        onError: (error) => {
          // Suppress the environments error during instance deletion
          if (error.message.includes('Expected Iterable') && error.message.includes('environments')) {
            console.warn('Suppressed environments GraphQL error in TemplatesTableLogic');
            // window.location.reload();
            return;
          }
          makeErrorCatcher(ErrorTypes.GenericError)(error);
        },
        document: updatedOwnedInstances,
        variables: {
          tenantNamespace,
        },
        updateQuery: updateQueryOwnedInstancesQuery(
          setDataInstances,
          userId ?? '',
          notifier,
        ),
      });
      return unsubscribe;
    }
  }, [
    loadingInstances,
    errorInstances,
    errorsQueue.length,
    subscribeToMoreInstances,
    tenantNamespace,
    userId,
    makeErrorCatcher,
    notifier,
  ]);

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
          onError: (error) => {
            // Suppress the environmentList error during template deletion
            if (error.message.includes('Expected Iterable') && error.message.includes('environmentList')) {
              console.warn('Suppressed environmentList GraphQL error in TemplatesTableLogic');
              // window.location.reload();
              return;
            }
            makeErrorCatcher(ErrorTypes.GenericError)(error);
          },
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
    })
      .then(i => {
        setDataInstances(old =>
          !old.find(x => x.name === i.data?.createdInstance?.metadata?.name)
            ? [
                ...old,
                makeGuiInstance(i.data?.createdInstance, userId, {
                  templateName: templateId,
                  workspaceName: workspaceName,
                }),
              ]
            : old,
        );
        // Refresh quota after instance creation
        refreshQuota?.();
        return i;
      })
      .catch(error => {
        console.error('TemplatesTableLogic createInstance error:', error);
        throw error;
      });

  const templates = useMemo(() => {
    const joined = joinInstancesAndTemplates(dataTemplate, dataInstances);

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
  }, [dataTemplate, dataInstances, templateListData?.templateList?.templates]);

  return (
    <>
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
          spinning={loadingTemplate || loadingInstances}
          style={{
            display: 'flex',
            flexDirection: 'column',
            flex: '1 1 auto',
            minHeight: 0,
          }}
        >
          {!loadingTemplate &&
          !loadingInstances &&
          !errorTemplate &&
          !errorInstances &&
          templates &&
          dataInstances ? (
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                flex: '1 1 auto',
                minHeight: 0,
              }}
            >
              <TemplatesTable
                totalInstances={dataInstances.length}
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
                  }).then(result => {
                    // Refresh quota after template deletion
                    refreshQuota?.();
                    return result;
                  })
                }
                deleteTemplateLoading={loadingDeleteTemplateMutation}
                editTemplate={() => null}
                createInstance={createInstance}
                availableQuota={availableQuota}
                refreshQuota={refreshQuota}
                isPersonal={isPersonal}
              />
            </div>
          ) : (
            <div
              className={
                loadingTemplate ||
                loadingInstances ||
                errorTemplate ||
                errorInstances
                  ? 'invisible'
                  : 'visible'
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

          {role === WorkspaceRole.manager &&
          !loadingTemplate &&
          !loadingInstances &&
          !isPersonal ? (
            <>
              <SharedVolumesDrawer
                workspaceNamespace={workspaceNamespace}
                isPersonal={isPersonal}
              />
            </>
          ) : null}
        </Spin>
      </div>
    </>
  );
};

export default TemplatesTableLogic;
