import { type FC, useContext, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { VncScreen, type VncScreenHandle } from 'react-vnc';
import { AuthContext } from '../../../contexts/AuthContext';
import { OwnedInstancesContext } from '../../../contexts/OwnedInstancesContext';
import './NativeVNCPage.css';

const NativeVNCPage: FC = () => {
  const { namespace = '', VMname = '', environment = '' } = useParams();
  const { instances, loading } = useContext(OwnedInstancesContext);
  const { token } = useContext(AuthContext);

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
  const vncRef = useRef<VncScreenHandle>(null);

  // On Chromium, clicking the screen makes noVNC show a mouse capture element, which is a div that covers the entire screen and captures all mouse events. 
  // This is necessary for noVNC to work properly, but it also means that the VncScreen component loses focus. 
  // If the user then clicks outside of the VncScreen component, it will lose focus and noVNC will stop capturing mouse events. 
  // This function checks if the mouse capture element is visible, and if so, it refocuses the VncScreen component.
  // Keep the focus while a capture is in progress (https://github.com/roerohan/react-vnc/issues/5).
  const keepFocusDuringCapture = () => {
    const capture = document.getElementById('noVNC_mouse_capture_elem');
    if (capture && capture.style.display !== 'none') {
      vncRef.current?.focus();
    }
  };

  if (loading) return <div className="native-vnc-page-status">Loading…</div>;

  if (!wsUrl || !token)
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
    <div onBlur={keepFocusDuringCapture}>
    <VncScreen
      ref={vncRef}
      url={wsUrl}
      // The gateway authenticates this endpoint with an OIDC redirect, which a WebSocket
      // handshake cannot follow, and the WebSocket API cannot set request headers. The
      // token therefore travels as the last WebSocket subprotocol, one of the locations
      // the gateway's JWT provider reads.
      // "binary" must come first, since QEMU rejects a handshake that does not offer it,
      // and the token last, since the gateway reads to the end of the header value.
      rfbOptions={{ wsProtocols: ['binary', token] }}
      resizeSession
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
    </div>
  );
};

export default NativeVNCPage;
