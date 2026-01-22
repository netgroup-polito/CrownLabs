import { useState, useCallback, useContext, useEffect, useRef } from 'react';
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
} from '../env';

export const useMydrive = () => {
  const [isLoading, setIsLoading] = useState(false);
  const isStartingRef = useRef(false);
  const [hasLoadedOnce, setHasLoadedOnce] = useState(false);

  const { data: tenantData } = useContext(TenantContext);
  const {
    instances: ownedInstances,
    loading: instancesLoading,
    data: instancesData,
  } = useContext(OwnedInstancesContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);

  // Track when data has been loaded at least once
  useEffect(() => {
    if (!hasLoadedOnce && !instancesLoading && instancesData) {
      setHasLoadedOnce(true);
    }
  }, [instancesLoading, instancesData, hasLoadedOnce]);

  const tenantId = tenantData?.tenant?.metadata?.name;
  const tenantNamespace = tenantData?.tenant?.status?.personalNamespace?.name;

  // Derive utilities workspace namespace dynamically from the workspace object
  const utilitiesWorkspace = tenantData?.tenant?.spec?.workspaces?.find(
    ws => ws?.name === VITE_APP_MYDRIVE_WORKSPACE_NAME,
  );

  const workspaceNamespace =
    utilitiesWorkspace?.workspaceWrapperTenantV1alpha2
      ?.itPolitoCrownlabsV1alpha1Workspace?.status?.namespace?.name;

  // Check if user has access to utilities workspace
  const hasUtilitiesAccess = Boolean(
    tenantData?.tenant?.spec?.workspaces?.some(
      ws => ws?.name === VITE_APP_MYDRIVE_WORKSPACE_NAME,
    ),
  );

  const [createInstanceMutation] = useCreateInstanceMutation({
    onError: apolloErrorCatcher,
  });

  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });

  // Find mydrive instance directly from ownedInstances (no state, computed value)
  // This ensures we always have the latest value without race conditions
  const mydriveInstance =
    ownedInstances.find(
      (inst: Instance) =>
        inst.templateName === VITE_APP_MYDRIVE_TEMPLATE_NAME &&
        inst.workspaceName === VITE_APP_MYDRIVE_WORKSPACE_NAME,
    ) || null;

  // Get the drive URL without opening it
  const getDriveUrl = useCallback((): string | null => {
    if (mydriveInstance?.status === Phase2.Ready && mydriveInstance.url) {
      const env = mydriveInstance.environments?.[0];
      if (env) {
        const baseUrl = mydriveInstance.url.endsWith('/')
          ? mydriveInstance.url.slice(0, -1)
          : mydriveInstance.url;
        return `${baseUrl}/${env.name}/`;
      }
    }
    return null;
  }, [mydriveInstance]);

  // Set prettyname if not already set
  useEffect(() => {
    if (mydriveInstance && mydriveInstance.prettyName !== 'filemanager') {
      setInstancePrettyname(
        'filemanager',
        mydriveInstance,
        applyInstanceMutation,
      );
    }
  }, [mydriveInstance, applyInstanceMutation]);

  // Monitor instance status - just stop loading when ready
  useEffect(() => {
    if (isLoading && mydriveInstance?.status === Phase2.Ready) {
      setIsLoading(false);
    }
  }, [mydriveInstance, isLoading]);

  // Start the drive instance without navigating (used by DriveView component)
  const startDriveInstance = useCallback(async () => {
    // Prevent concurrent calls
    if (isStartingRef.current) {
      return;
    }

    if (!tenantId || !tenantNamespace) {
      console.error('Tenant information not available');
      return;
    }

    if (!workspaceNamespace) {
      console.error('Utilities workspace namespace not available');
      return;
    }

    // If instance is ready or already starting/running, no need to do anything
    if (
      mydriveInstance?.status === Phase2.Ready ||
      mydriveInstance?.status === Phase2.Starting ||
      mydriveInstance?.status === Phase2.Importing
    ) {
      return;
    }

    isStartingRef.current = true;
    setIsLoading(true);

    try {
      if (!mydriveInstance) {
        // Create new instance - Kubernetes will auto-generate the name
        await createInstanceMutation({
          variables: {
            templateId: VITE_APP_MYDRIVE_TEMPLATE_NAME,
            tenantNamespace,
            tenantId,
            workspaceNamespace,
          },
        });
      } else if (mydriveInstance.status === Phase2.Off) {
        // Instance is off, start it
        await setInstanceRunning(true, mydriveInstance, applyInstanceMutation);
      }
      // For other states (starting, etc.), just wait for it to be ready
    } catch (error) {
      console.error('Error starting mydrive instance:', error);
      isStartingRef.current = false;
      setIsLoading(false);
    }
  }, [
    tenantId,
    tenantNamespace,
    workspaceNamespace,
    mydriveInstance,
    createInstanceMutation,
    applyInstanceMutation,
  ]);

  // Reset starting ref when instance becomes ready
  useEffect(() => {
    if (mydriveInstance?.status === Phase2.Ready) {
      isStartingRef.current = false;
    }
  }, [mydriveInstance?.status]);

  // instancesLoaded is true only when data has been loaded at least once
  // This prevents creating instances before we know if one already exists
  const instancesLoaded = hasLoadedOnce;

  return {
    startDriveInstance,
    isLoading,
    mydriveInstance,
    hasUtilitiesAccess,
    getDriveUrl,
    instancesLoaded,
  };
};
