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
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);
  const { tenantNamespace, workspaceNamespace, workspaceName, role } = props;

  const [dataInstances, setDataInstances] = useState<Instance[]>([]);

  const notifier = useContext(TenantContext).notify;

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
    onError: apolloErrorCatcher,
    onCompleted: data =>
      setDataInstances(
        data.instanceList?.instances
          ?.map(i => makeGuiInstance(i, userId))
          .sort((a, b) =>
            (a.prettyName ?? '').localeCompare(b.prettyName ?? ''),
          ) ?? [],
      ),
    fetchPolicy: fetchPolicy_networkOnly,
  });

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

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
    () =>
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
        .sort((a, b) => a.name.localeCompare(b.name)) ?? [],
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
    () => joinInstancesAndTemplates(dataTemplate, dataInstances),
    [dataTemplate, dataInstances],
  );

  return (
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
  );
};

export default TemplatesTableLogic;
