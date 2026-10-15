import {
  type FC,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
} from 'react';
import {
  type UpdatedWorkspaceImagesSubscription,
  useWorkspaceImagesQuery,
} from '../generated-types';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import { updatedWorkspaceImages } from '../graphql-components/subscription';
import { updateWorkspaceImages } from '../components/workspaces/Images/workspaceImagesUpdates';
import { WorkspaceImagesContext } from './WorkspaceImagesContext';

interface WorkspaceImagesContextProviderProps extends PropsWithChildren {
  workspaceNamespace: string;
}

const WorkspaceImagesContextProvider: FC<WorkspaceImagesContextProviderProps> = ({
  children,
  workspaceNamespace,
}) => {
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const {
    data,
    loading,
    error,
    subscribeToMore,
  } = useWorkspaceImagesQuery({
    variables: { workspaceNamespace },
    skip: !workspaceNamespace,
    // Read a cached value immediately, then refresh it when the workspace opens.
    fetchPolicy: 'cache-and-network',
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (!workspaceNamespace || loading || error) return;

    const unsubscribe = subscribeToMore<UpdatedWorkspaceImagesSubscription>({
      document: updatedWorkspaceImages,
      variables: { workspaceNamespace },
      onError: makeErrorCatcher(ErrorTypes.GenericError),
      updateQuery: (previous, { subscriptionData }) =>
        updateWorkspaceImages(
          previous,
          subscriptionData.data?.updatedImage,
          workspaceNamespace,
        ),
    });

    return () => unsubscribe();
  }, [
    error,
    loading,
    makeErrorCatcher,
    subscribeToMore,
    workspaceNamespace,
  ]);

  const value = useMemo(
    () => ({
      data,
      loading,
      error: error ? new Error(error.message) : undefined,
      workspaceNamespace,
    }),
    [data, error, loading, workspaceNamespace],
  );

  return (
    <WorkspaceImagesContext.Provider value={value}>
      {children}
    </WorkspaceImagesContext.Provider>
  );
};

export default WorkspaceImagesContextProvider;
