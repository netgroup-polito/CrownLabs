import {
  type FC,
  type KeyboardEvent,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react';
import { useParams } from 'react-router-dom';
import { VncScreen, type VncScreenHandle } from 'react-vnc';
import { AuthContext } from '../../../contexts/AuthContext';
import './NativeVNCPage.css';

// X11 keysym of the V key.
const XK_v = 0x0076;
// Formats of a Kubernetes UID and of an environment name.
const UID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const ENV_RE = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;

const NativeVNCPage: FC = () => {
  const { instanceId = '', environment = '' } = useParams();
  const { token } = useContext(AuthContext);

  // The link picks only the instance and its environment. The host is always the page's
  // own, so the token sent with the connection never leaves CrownLabs.
  const wsUrl =
    UID_RE.test(instanceId) && ENV_RE.test(environment)
      ? `wss://${window.location.host}/instance/${instanceId}/${environment}/`
      : undefined;

  const [connectionFailed, setConnectionFailed] = useState(false);
  const connectedRef = useRef(false);
  const vncRef = useRef<VncScreenHandle>(null);

  // Keep the focus while a capture is in progress (https://github.com/roerohan/react-vnc/issues/5).
  const keepFocusDuringCapture = () => {
    const capture = document.getElementById('noVNC_mouse_capture_elem');
    if (capture && capture.style.display !== 'none') {
      vncRef.current?.focus();
    }
  };

  // Last text synced between the local and the VM clipboard, in either direction, so that
  // a text is never sent back to the side it came from.
  const lastClipboardTextRef = useRef<string | null>(null);
  // Whether Ctrl+V takes the local clipboard from the browser's paste event, since the
  // browser cannot read it when the console gets the focus.
  const pasteFromEventRef = useRef(false);
  // Physical key of a Ctrl+V waiting for the browser's paste event.
  const pendingPasteKeyRef = useRef<string | null>(null);

  // The VM clipboard changed: copy its text to the local clipboard.
  const copyToLocalClipboard = (e: CustomEvent<{ text: string }>) => {
    lastClipboardTextRef.current = e.detail.text;
    // Rejected when the page has no focus (Chromium) or no recent user input (Firefox).
    navigator.clipboard.writeText(e.detail.text).catch(() => {});
  };

  // The console got the focus: send the local clipboard to the VM, ready to be pasted there.
  const sendLocalClipboard = async () => {
    const permission = await navigator.permissions
      .query({ name: 'clipboard-read' as PermissionName })
      .catch(() => null);
    // Not supported by the browser (Firefox, Safari) or denied by the user.
    pasteFromEventRef.current = !permission || permission.state === 'denied';
    if (pasteFromEventRef.current || !connectedRef.current) return;
    try {
      const text = await navigator.clipboard.readText();
      if (text && text !== lastClipboardTextRef.current) {
        lastClipboardTextRef.current = text;
        vncRef.current?.clipboardPaste(text);
      }
    } catch {
      // Not allowed by the user, or the page lost the focus meanwhile.
    }
  };

  // Ctrl+V, also with Shift, when the text comes from the paste event: keep the key from
  // noVNC, which would block it, so that the browser runs its own paste.
  const letBrowserPaste = (e: KeyboardEvent) => {
    if (!pasteFromEventRef.current) return;
    if (!e.ctrlKey || e.altKey || e.key.toLowerCase() !== 'v') return;
    e.stopPropagation();
    pendingPasteKeyRef.current = e.code;
  };

  // The browser ran the paste of a Ctrl+V: send its text to the VM clipboard, then press
  // V in the VM, where Ctrl (and Shift) are still held down, so that the VM pastes it.
  // The paste event targets the selection or the body, not the console: hence the
  // listener on the document.
  useEffect(() => {
    const pasteToVM = (e: ClipboardEvent) => {
      const code = pendingPasteKeyRef.current;
      if (!code) return;
      pendingPasteKeyRef.current = null;
      e.preventDefault();
      const text = e.clipboardData?.getData('text/plain');
      if (text && text !== lastClipboardTextRef.current) {
        lastClipboardTextRef.current = text;
        vncRef.current?.clipboardPaste(text);
      }
      vncRef.current?.sendKey(XK_v, code);
    };
    document.addEventListener('paste', pasteToVM);
    return () => document.removeEventListener('paste', pasteToVM);
  }, []);

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
    <div onBlur={keepFocusDuringCapture} onFocus={sendLocalClipboard} onKeyDownCapture={letBrowserPaste}>
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
      onClipboard={copyToLocalClipboard}
    />
    </div>
  );
};

export default NativeVNCPage;
