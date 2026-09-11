import { type FC } from 'react';
import { VncScreen } from 'react-vnc';
import './NativeVNCViewer.css';

export interface INativeVNCViewerProps {
  wsUrl: string;
}

const NativeVNCViewer: FC<INativeVNCViewerProps> = ({ wsUrl }) => {
  return (
    <VncScreen
      url={wsUrl}
      scaleViewport
      focusOnClick
      background="#000000"
      className="native-vnc-viewer"
      onConnect={() => console.log('[native-vnc] connected')}
      onDisconnect={() => console.log('[native-vnc] disconnected')}
    />
  );
};

export default NativeVNCViewer;
