/* eslint-disable @typescript-eslint/no-unused-vars */
import { FetchPolicy } from '@apollo/client';
import { Spin } from 'antd';
import { useContext, useEffect, useState } from 'react';
import { FC } from 'react';
import { AuthContext } from '../../../../contexts/AuthContext';
import {
  useCreateInstanceMutation,
  useDeleteTemplateMutation,
  useOwnedInstancesQuery,
  useWorkspaceTemplatesQuery,
} from '../../../../generated-types';
import {
  updatedOwnedInstances,
  updatedWorkspaceTemplates,
} from '../../../../graphql-components/subscription';
import { Instance, Template, WorkspaceRole } from '../../../../utils';
import {
  getInstance,
  getTemplate,
  joinInstancesAndTemplates,
  updateQueryOwnedInstancesQuery,
  updateQueryWorkspaceTemplatesQuery,
} from '../../../../utilsLogic';
import { TemplatesEmpty } from '../TemplatesEmpty';
import { TemplatesTable } from '../TemplatesTable';

export interface ITemplateTableLogicProps {
  tenantNamespace: string;
  workspaceNamespace: string;
  role: WorkspaceRole;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';

const TemplatesTableLogic: FC<ITemplateTableLogicProps> = ({ ...props }) => {
  const { userId } = useContext(AuthContext);
  const { tenantNamespace, workspaceNamespace, role } = props;

  const [dataInstances, setDataInstances] = useState<Instance[]>([]);

  const {
    loading: loadingInstances,
    error: errorInstances,
    subscribeToMore: subscribeToMoreInstances,
  } = useOwnedInstancesQuery({
    variables: { tenantNamespace },
    onCompleted: data =>
      setDataInstances(
        data.instanceList?.instances?.map((i, n) =>
          getInstance(i!, n, tenantNamespace, workspaceNamespace)
        )!
      ),
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    if (!loadingInstances) {
      subscribeToMoreInstances({
        document: updatedOwnedInstances,
        variables: {
          tenantNamespace,
        },
        updateQuery: updateQueryOwnedInstancesQuery(
          dataInstances,
          setDataInstances,
          tenantNamespace,
          workspaceNamespace
        ),
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingInstances, subscribeToMoreInstances, tenantNamespace, userId]);

  const [dataTemplate, setDataTemplate] = useState<Template[]>([]);

  const {
    loading: loadingTemplate,
    error: errorTemplate,
    subscribeToMore: subscribeToMoreTemplates,
  } = useWorkspaceTemplatesQuery({
    variables: { workspaceNamespace },
    onCompleted: data =>
      setDataTemplate(
        data.templateList?.templates?.map(t =>
          getTemplate({
            original: t!,
            alias: {
              id: t?.metadata?.id!,
              name: t?.spec?.name!,
            },
          })
        )!
      ),
    fetchPolicy: fetchPolicy_networkOnly,
  });

  useEffect(() => {
    if (!loadingTemplate) {
      subscribeToMoreTemplates({
        document: updatedWorkspaceTemplates,
        variables: { workspaceNamespace: `${workspaceNamespace}` },
        updateQuery: updateQueryWorkspaceTemplatesQuery(setDataTemplate),
      });
    }
  }, [loadingTemplate, subscribeToMoreTemplates, userId, workspaceNamespace]);

  const [createInstanceMutation] = useCreateInstanceMutation();
  const [deleteTemplateMutation, { loading: loadingDeleteTemplateMutation }] =
    useDeleteTemplateMutation();

  const deleteTemplate = (templateId: string) =>
    deleteTemplateMutation({
      variables: {
        workspaceNamespace,
        templateId,
      },
    });

  const createInstance = (templateId: string) =>
    createInstanceMutation({
      variables: {
        templateId,
        tenantNamespace,
        tenantId: userId!,
        workspaceNamespace,
      },
    }).then(i => {
      setDataInstances(old => [
        ...old,
        getInstance(
          i.data?.createdInstance!,
          old.length,
          tenantNamespace,
          workspaceNamespace
        ),
      ]);
      return i;
    });

  const templates = joinInstancesAndTemplates(dataTemplate, dataInstances);

  return (
    <Spin size="large" spinning={loadingTemplate || loadingInstances}>
      {!loadingTemplate &&
      !loadingInstances &&
      !errorTemplate &&
      !errorInstances &&
      dataInstances &&
      templates &&
      dataInstances ? (
        <TemplatesTable
          tenantNamespace={tenantNamespace}
          workspaceNamespace={workspaceNamespace}
          templates={templates}
          role={role}
          deleteTemplate={deleteTemplate}
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
    </Spin>
  );
};

export default TemplatesTableLogic;
