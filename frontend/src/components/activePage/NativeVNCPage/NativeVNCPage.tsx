import { type FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { VncScreen } from 'react-vnc';
import { useInstanceStatusQuery } from '../../../generated-types';
import './NativeVNCPage.css';

const NativeVNCPage: FC = () => {
  const {
    namespace = '',
    VMname: name = '',
    environment = '',
  } = useParams();

  const { data, loading, error } = useInstanceStatusQuery({
    variables: { name, namespace },
  });

  const instanceUrl = data?.instance?.status?.url;

  const wsUrl = (() => {
    if (!instanceUrl) return undefined;
    const baseUrl = instanceUrl.endsWith('/')
      ? instanceUrl.slice(0, -1)
      : instanceUrl;
    return `${baseUrl}/${environment}/`.replace(/^https/, 'wss');
  })();

  const [sessionState, setSessionState] = useState<
    'checking' | 'ready' | 'needsLogin'
  >('checking');

  useEffect(() => {
    if (!instanceUrl) return;
    let cancelled = false;

    fetch(instanceUrl, { credentials: 'include' })
      .then(res => {
        if (cancelled) return;
        const backHome =
          new URL(res.url).origin === new URL(instanceUrl).origin;
        setSessionState(backHome ? 'ready' : 'needsLogin');
      })
      .catch(() => {
        if (!cancelled) setSessionState('ready');
      });

    return () => {
      cancelled = true;
    };
  }, [instanceUrl]);

  if (loading || sessionState === 'checking')
    return <div className="native-vnc-page-status">Loading…</div>;

  if (error || !wsUrl)
    return (
      <div className="native-vnc-page-status">
        Unable to load the VNC connection for this instance.
      </div>
    );

  if (sessionState === 'needsLogin')
    return (
      <div className="native-vnc-page-status">
        <div>
          <p>Your session has expired.</p>
          <button onClick={() => instanceUrl && window.open(instanceUrl, '_blank')}>
            Log in again
          </button>
        </div>
      </div>
    );

  return (
    <VncScreen
      url={wsUrl}
      scaleViewport
      focusOnClick
      background="#000000"
      className="native-vnc-page"
      onConnect={() => console.log('[native-vnc] connected')}
      onDisconnect={() => console.log('[native-vnc] disconnected')}
    />
  );
};

export default NativeVNCPage;
