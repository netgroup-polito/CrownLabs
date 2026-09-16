import { Button } from 'antd';
import { type FC, useContext, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { VncScreen } from 'react-vnc';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import './NativeVNCPage.css';

type SessionState = 'checking' | 'ready' | 'loginRequired';

const NativeVNCPage: FC = () => {
  const { namespace = '', VMname = '', environment = '' } = useParams();
  const { instances, loading } = useContext(OwnedInstancesContext);

  const instance = instances.find(
    i => i.name === VMname && i.tenantNamespace === namespace,
  );

  const environmentUrl = (() => {
    if (!instance?.url) return undefined;
    const baseUrl = instance.url.endsWith('/')
      ? instance.url.slice(0, -1)
      : instance.url;
    return `${baseUrl}/${environment}/`;
  })();

  const [sessionState, setSessionState] = useState<SessionState>('checking');
  const [attempt, setAttempt] = useState(0);
  const [connectionFailed, setConnectionFailed] = useState(false);
  const connectedRef = useRef(false);

  // The instance endpoint is protected by a per-instance OIDC cookie gate that a
  // WebSocket handshake cannot satisfy on its own: browsers do not follow
  // redirects for WebSockets and report the failure as an opaque 1006 close.
  // Loading the endpoint in a hidden iframe first lets the browser walk the
  // whole OIDC redirect chain and store the session cookies. If the chain ends
  // back on our own origin the session is established; if it stops on the
  // identity provider the document is cross-origin, reading its location throws,
  // and the user has to authenticate interactively.
  const handleSessionProbeLoad = (
    event: React.SyntheticEvent<HTMLIFrameElement>,
  ) => {
    try {
      const href = event.currentTarget.contentWindow?.location.href;
      setSessionState(href ? 'ready' : 'loginRequired');
    } catch {
      setSessionState('loginRequired');
    }
  };

  if (loading) return <div className="native-vnc-page-status">Loading…</div>;

  if (!environmentUrl)
    return (
      <div className="native-vnc-page-status">
        Unable to load the VNC connection for this instance.
      </div>
    );

  if (sessionState === 'checking')
    return (
      <div className="native-vnc-page-status">
        Establishing session…
        <iframe
          key={attempt}
          title="VNC session"
          src={environmentUrl}
          onLoad={handleSessionProbeLoad}
          className="native-vnc-session-probe"
        />
      </div>
    );

  if (sessionState === 'loginRequired')
    return (
      <div className="native-vnc-page-status">
        <span>Your session for this instance has expired.</span>
        <a href={environmentUrl} target="_blank" rel="noreferrer">
          Sign in again
        </a>
        <Button
          shape="round"
          onClick={() => {
            setSessionState('checking');
            setAttempt(a => a + 1);
          }}
        >
          Retry
        </Button>
      </div>
    );

  if (connectionFailed)
    return (
      <div className="native-vnc-page-status">
        Connection to the VNC server failed.
      </div>
    );

  return (
    <VncScreen
      url={environmentUrl.replace(/^https/, 'wss')}
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
