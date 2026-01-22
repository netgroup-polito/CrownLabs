import { useState, useCallback, useContext, useEffect } from 'react';
import { Modal } from 'antd';
import { TenantContext } from '../contexts/TenantContext';
import { OwnedInstancesContext } from '../contexts/OwnedInstancesContext';
import { ErrorContext } from '../errorHandling/ErrorContext';
import {
  useCreateInstanceMutation,
  useApplyInstanceMutation,
  Phase2,
} from '../generated-types';
import type { Instance } from '../utils';
import { setInstanceRunning, setInstancePrettyname } from '../utilsLogic';
import {
  VITE_APP_MYDRIVE_TEMPLATE_NAME,
  VITE_APP_MYDRIVE_WORKSPACE_NAME,
  VITE_APP_MYDRIVE_WORKSPACE_NAMESPACE,
} from '../env';

export const useMydrive = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [mydriveInstance, setMydriveInstance] = useState<Instance | null>(null);

  const { data: tenantData } = useContext(TenantContext);
  const { instances: ownedInstances } = useContext(OwnedInstancesContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const tenantId = tenantData?.tenant?.metadata?.name;
  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  // Check if user has access to utilities workspace
  const hasUtilitiesAccess = Boolean(
    tenantData?.tenant?.spec?.workspaces?.some(
      ws => ws?.name === VITE_APP_MYDRIVE_WORKSPACE_NAME
    )
  );

  const [createInstanceMutation] = useCreateInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });

  // Find mydrive instance in owned instances
  // We look for any instance with the mydrive template in utilities workspace
  // Since Kubernetes generates the name automatically, we can't rely on a fixed name
  useEffect(() => {
    const instance = ownedInstances.find(
      (inst: Instance) =>
        inst.templateName === VITE_APP_MYDRIVE_TEMPLATE_NAME &&
        inst.workspaceName === VITE_APP_MYDRIVE_WORKSPACE_NAME,
    );
    setMydriveInstance(instance || null);
  }, [ownedInstances]);

  // Check if instance is ready and open it
  const openMydrive = useCallback(() => {
    if (mydriveInstance?.status === Phase2.Ready && mydriveInstance.url) {
      const env = mydriveInstance.environments?.[0];
      if (env) {
        const baseUrl = mydriveInstance.url.endsWith('/')
          ? mydriveInstance.url.slice(0, -1)
          : mydriveInstance.url;
        const envUrl = `${baseUrl}/${env.name}/`;
        window.open(envUrl, '_blank');
        return true;
      }
    }
    return false;
  }, [mydriveInstance]);

  // Set prettyname if not already set
  useEffect(() => {
    if (mydriveInstance && mydriveInstance.prettyName !== 'filemanager') {
      setInstancePrettyname('filemanager', mydriveInstance, applyInstanceMutation);
    }
  }, [mydriveInstance, applyInstanceMutation]);

  // Monitor instance status and open when ready
  useEffect(() => {
    if (isLoading && mydriveInstance?.status === Phase2.Ready) {
      setIsLoading(false);
      openMydrive();
    }
  }, [mydriveInstance, isLoading, openMydrive]);

  const handleDriveClick = useCallback(async () => {
    if (!tenantId || !tenantNamespace) {
      console.error('Tenant information not available');
      return;
    }

    setIsLoading(true);

    try {
      if (!mydriveInstance) {
        // Create new instance - Kubernetes will auto-generate the name
        await createInstanceMutation({
          variables: {
            templateId: VITE_APP_MYDRIVE_TEMPLATE_NAME,
            tenantNamespace,
            tenantId,
            workspaceNamespace: VITE_APP_MYDRIVE_WORKSPACE_NAMESPACE,
          },
        });
        // Show info modal
        Modal.info({
          title: 'Avvio in corso',
          content:
            "L'instanza per accedere al drive sta per essere avviata. Attendi qualche istante, verrà aperta automaticamente quando pronta.",
          okText: 'OK',
        });
        // Instance will be picked up by the useEffect and will open when ready
      } else {
        // Instance exists - check its status
        if (mydriveInstance.status === Phase2.Ready) {
          // Instance is ready, open it
          setIsLoading(false);
          openMydrive();
        } else if (mydriveInstance.status === Phase2.Off) {
          // Instance is off, start it
          await setInstanceRunning(true, mydriveInstance, applyInstanceMutation);
          // Show info modal
          Modal.info({
            title: 'Avvio in corso',
            content:
              "L'instanza per accedere al drive sta per essere avviata. Attendi qualche istante, verrà aperta automaticamente quando pronta.",
            okText: 'OK',
          });
          // Will be opened when ready by useEffect
        } else {
          // Instance is starting or in another transitional state, just wait
          Modal.info({
            title: 'Avvio in corso',
            content:
              "L'instanza per accedere al drive è già in fase di avvio. Attendi qualche istante, verrà aperta automaticamente quando pronta.",
            okText: 'OK',
          });
          // Will be opened when ready by useEffect
        }
      }
    } catch (error) {
      console.error('Error handling mydrive instance:', error);
      setIsLoading(false);
    }
  }, [
    tenantId,
    tenantNamespace,
    mydriveInstance,
    createInstanceMutation,
    applyInstanceMutation,
    openMydrive,
  ]);

  return {
    handleDriveClick,
    isLoading,
    mydriveInstance,
    hasUtilitiesAccess,
  };
};
