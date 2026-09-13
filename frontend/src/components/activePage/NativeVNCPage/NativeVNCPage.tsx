import { type FC, useContext, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { VncScreen } from 'react-vnc';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import './NativeVNCPage.css';

const NativeVNCPage: FC = () => {
  const { namespace = '', VMname = '', environment = '' } = useParams();
  const { instances, loading } = useContext(OwnedInstancesContext);

  const instance = instances.find(
    i => i.name === VMname && i.tenantNamespace === namespace,
  );

  const wsUrl = (() => {
    if (!instance?.url) return undefined;
    const baseUrl = instance.url.endsWith('/')
      ? instance.url.slice(0, -1)
      : instance.url;
    return `${baseUrl}/${environment}/`.replace(/^https/, 'wss');
  })();

  const [connectionFailed, setConnectionFailed] = useState(false);
  const connectedRef = useRef(false);

  if (loading) return <div className="native-vnc-page-status">Loading…</div>;

  if (!wsUrl)
    return (
      <div className="native-vnc-page-status">
        Unable to load the VNC connection for this instance.
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
