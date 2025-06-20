import {
  type FC,
  type PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useState,
} from 'react';
import { ErrorContext } from '../errorHandling/ErrorContext';
import { ErrorTypes } from '../errorHandling/utils';
import {
  useApplyTenantMutation,
  useTenantQuery,
  type UpdatedTenantSubscription,
} from '../generated-types';
import { updatedTenant } from '../graphql-components/subscription';
import { getTenantPatchJson } from '../graphql-components/utils';
import { TenantContext } from './TenantContext';
import { AuthContext } from './AuthContext';
import { message } from 'antd';
import type { JointContent } from 'antd/lib/message/interface';

const TenantContextProvider: FC<PropsWithChildren> = props => {
  const { userId } = useContext(AuthContext);
  const { children } = props;

  const [now, setNow] = useState(new Date());

  const { profile } = useContext(AuthContext);

  const { makeErrorCatcher, apolloErrorCatcher, errorsQueue } =
    useContext(ErrorContext);

  const { loading, error, subscribeToMore, data } = useTenantQuery({
    skip: !userId,
    variables: { tenantId: userId || '' },
    fetchPolicy: 'network-only',
    nextFetchPolicy: 'cache-only',
    onError: apolloErrorCatcher,
  });

  const [applyTenantMutation] = useApplyTenantMutation({
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (userId && !loading && !error && !errorsQueue.length) {
      const unsubscribe = subscribeToMore<UpdatedTenantSubscription>({
        onError: makeErrorCatcher(ErrorTypes.ApolloError),
        variables: { tenantId: userId ?? '' },
        document: updatedTenant,
        updateQuery: (prev, { subscriptionData }) => {
          if (!subscriptionData.data.updatedTenant?.tenant) return prev;

          return Object.assign({}, prev, {
            tenant: subscriptionData.data.updatedTenant.tenant,
          });
        },
      });
      return unsubscribe;
    }
  }, [
    subscribeToMore,
    loading,
    userId,
    errorsQueue.length,
    error,
    apolloErrorCatcher,
    makeErrorCatcher,
  ]);

  const refreshClock = () => setNow(new Date());

  useEffect(() => {
    const timerHandler = setInterval(refreshClock, 60000);
    return () => clearInterval(timerHandler);
  }, []);

  const patchTenantLastLogin = useCallback((tenantId: string) => {
    applyTenantMutation({
      variables: {
        tenantId,
        patchJson: getTenantPatchJson({
          lastLogin: new Date(),
        }),
        manager: 'frontend-tenant-lastlogin',
      },
      onError: apolloErrorCatcher,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!data?.tenant?.metadata?.name || !userId) return;
    patchTenantLastLogin(userId);
  }, [userId, data?.tenant?.metadata?.name, patchTenantLastLogin]);

  const [messageApi, contextHolder] = message.useMessage();

  const [messages, setMessages] = useState<
    {
      type: 'warning' | 'success';
      key: string;
      content: JointContent;
    }[]
  >([]);

  useEffect(() => {
    const [msg] = messages;
    if (!msg) return;
    messageApi[msg.type](msg.content);
    setMessages(messages => messages.filter(m => m.key !== msg.key));
  }, [messageApi, messages, setMessages]);

  const notify = useCallback(
    (type: 'warning' | 'success', key: string, content: JointContent) => {
      setMessages(messages => {
        if (messages.find(m => m.key === key)) {
          return messages;
        }
        return [...messages, { type, key, content }];
      });
    },
    [setMessages],
  );

  const tSpec = data?.tenant?.spec;
  const displayName = tSpec
    ? `${tSpec.firstName} ${tSpec.lastName}`
    : profile?.name || '';

  const hasSSHKeys = !!tSpec?.publicKeys?.length;
  return (
    <TenantContext.Provider
      value={{
        data,
        loading,
        error,
        now,
        refreshClock,
        hasSSHKeys,
        displayName,
        notify,
      }}
    >
      {contextHolder}
      {children}
    </TenantContext.Provider>
  );
};

export default TenantContextProvider;
