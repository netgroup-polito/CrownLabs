import { type FC, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Button } from 'antd';
import { VncScreen } from 'react-vnc';
import { useInstanceStatusQuery } from '../../../generated-types';
import { refreshInstanceSession } from '../../../utils/nativeVncSession';
import './NativeVNCPage.css';

const NativeVNCPage: FC = () => {
  const { namespace = '', VMname: name = '', environment = '' } = useParams();

  const { data, loading, error } = useInstanceStatusQuery({
    variables: { name, namespace },
  });

  const instanceUrl = data?.instance?.status?.url;
  const envUrl = (() => {
    if (!instanceUrl) return undefined;
    const baseUrl = instanceUrl.endsWith('/')
      ? instanceUrl.slice(0, -1)
      : instanceUrl;
    return `${baseUrl}/${environment}/`;
  })();
  const wsUrl = envUrl?.replace(/^https/, 'wss');

  const [connectionFailed, setConnectionFailed] = useState(false);
  const connectedRef = useRef(false);

  if (loading) return <div className="native-vnc-page-status">Loading…</div>;

  if (error || !envUrl || !wsUrl)
    return (
      <div className="native-vnc-page-status">
        Unable to load the VNC connection for this instance.
      </div>
    );

  if (connectionFailed)
    return (
      <div className="native-vnc-page-status">
        <div>
          <p>Unable to connect. Your session may have expired.</p>
          <Button
            onClick={async () => {
              const sessionWindow = await refreshInstanceSession(envUrl);
              if (sessionWindow) {
                sessionWindow.close();
                window.location.reload();
              }
            }}
          >
            Log in again
          </Button>
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
      onConnect={() => {
        connectedRef.current = true;
      }}
      onDisconnect={() => {
        if (!connectedRef.current) setConnectionFailed(true);
      }}
    />
  );
};

export default NativeVNCPage;
