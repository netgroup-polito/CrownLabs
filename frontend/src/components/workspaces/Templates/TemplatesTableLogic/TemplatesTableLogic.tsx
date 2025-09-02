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
import QuotaDisplay from '../../QuotaDisplay/QuotaDisplay';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  workspaceName: string;
  role: WorkspaceRole;
  workspaceQuota: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
  isPersonal?: boolean;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';
const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  // const { userId } = useContext(AuthContext);
  const { user } = useContext(AuthContext);
  const userId = user?.profile?.sub;
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);
  const { tenantNamespace, workspaceNamespace, workspaceName, role, workspaceQuota, isPersonal } = props;

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
      const instances = data?.instanceList?.instances
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
  useEffect(() => {
    if (!loadingInstances && !errorInstances && !errorsQueue.length) {
      const unsubscribe = subscribeToMoreInstances({
        onError: makeErrorCatcher(ErrorTypes.GenericError),
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
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

  const notifier = useContext(TenantContext).notify;

  const {
    loading: loadingTemplate,
    error: errorTemplate,
    subscribeToMore: subscribeToMoreTemplates,
    data: templateListData,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    onError: apolloErrorCatcher,
    fetchPolicy: fetchPolicy_networkOnly,
    nextFetchPolicy: 'cache-only',
  });

  const dataTemplate = useMemo(
    () => {
      const templates = templateListData?.templateList?.templates
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
    },
    [templateListData?.templateList?.templates],
  );

  useEffect(() => {
    if (!loadingTemplate && !errorTemplate && !errorsQueue.length) {
      const unsubscribe =
        subscribeToMoreTemplates<UpdatedWorkspaceTemplatesSubscription>({
          onError: makeErrorCatcher(ErrorTypes.GenericError),
          document: updatedWorkspaceTemplates,
          variables: { workspaceNamespace },
          updateQuery: (prev, { subscriptionData }) => {
            const { data } = subscriptionData;
            if (!data?.updatedTemplate?.template) return prev;
            const { template, updateType } = data.updatedTemplate;
            const templates = prev.templateList?.templates ?? [];
            let out = [] as NonNullable<
              NonNullable<
                UpdatedWorkspaceTemplatesSubscriptionResult['data']
              >['updatedTemplate']
            >['template'][];
            switch (updateType) {
              case UpdateType.Added:
                out = [...templates, template];
                break;
              case UpdateType.Modified:
                out = templates.map(t =>
                  t?.metadata?.name === template.metadata?.name ? template : t,
                );
                break;
              case UpdateType.Deleted:
                out = templates.filter(
                  t => t?.metadata?.name !== template.metadata?.name,
                );
                break;
            }
            return Object.assign({}, prev, {
              templateList: {
                templates: out,
                __typename: prev.templateList?.__typename,
              },
            });
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

  const createInstance = (templateId: string, nodeSelector?: JSON) =>
    createInstanceMutation({
      variables: {
        templateId,
        tenantNamespace,
        tenantId: userId ?? '',
        workspaceNamespace,
        nodeSelector,
      },
    }).then(i => {
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
      return i;
    });

  const templates = useMemo(
    () => {      
      const joined = joinInstancesAndTemplates(dataTemplate, dataInstances);
      return joined;
    },
    [dataTemplate, dataInstances],
  );

  return (
    <>
      {isPersonal && (
        <QuotaDisplay
          tenantNamespace={tenantNamespace}
          templates={templates}
          instances={dataInstances}
          workspaceQuota={workspaceQuota}
        />
      )}
      <Spin size="large" spinning={loadingTemplate || loadingInstances}>
        {!loadingTemplate &&
        !loadingInstances &&
        !errorTemplate &&
        !errorInstances &&
        templates &&
        dataInstances ? (
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
              })
            }
            deleteTemplateLoading={loadingDeleteTemplateMutation}
            editTemplate={() => null}
            createInstance={createInstance}
            workspaceQuota={workspaceQuota}
            isPersonal={isPersonal}
          />
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
          >
            <TemplatesEmpty role={role} />
          </div>
        )}
        {role === WorkspaceRole.manager &&
        !loadingTemplate &&
        !loadingInstances ? (
          <SharedVolumesDrawer workspaceNamespace={workspaceNamespace} />
        ) : null}
      </Spin>
    </>
  );
};

export default TemplatesTableLogic;
