import { FC, useEffect, useRef, useState } from 'react';
import Clipboard from './clipboard';
// @ts-ignore (novnc has no TS declarations - yet)
import RFB from '@novnc/novnc/core/rfb.js';

export interface INovncProps {
  targetWS?: string;
  viewOnly?: boolean;
  quality?: number;
  compression?: number;
  username?: string;
  password?: string;
  onConnect?: () => void;
  onDisconnect?: () => void;
  errorHandler?: (e: Error) => void;
}

const NoVnc: FC<INovncProps> = ({ ...props }) => {
  const {
    targetWS,
    viewOnly = false,
    quality = 6,
    compression = 2,
    username = null,
    password = null,
    onConnect = () => {},
    onDisconnect = () => {},
    errorHandler = (...e) => console.error('NOVNC_ERROR', ...e),
  } = props;
  const ref = useRef<HTMLDivElement>(null);
  const [rfb, setRfb] = useState<RFB>(null);
  const [cb, setCb] = useState<Clipboard>();

  const disconnect = () => {
    rfb && rfb.disconnect();
    cb && cb.ungrab();
    setRfb(null);
    setCb(undefined);
    onDisconnect();
  };

  const connect = () => {
    if (!ref.current) {
      errorHandler(new Error('invalid_ref'));
      return;
    }
    const rfb = new RFB(ref.current, targetWS);

    rfb.addEventListener('connect', onConnect);
    rfb.addEventListener('disconnect', disconnect);
    rfb.addEventListener('credentialsrequired', () => {
      rfb.sendCredentials({ username, password });
    });
    rfb.addEventListener('securityfailure', () => {
      errorHandler(new Error('security_failure'));
    });

    rfb.scaleViewport = true;
    rfb.showDotCursor = true;
    rfb.background = `url(https://crownlabs.polito.it/error-page/crown.svg);`; //css value for crownlabs crown

    if (Clipboard.isSupported) {
      const cb = new Clipboard(ref.current);
      cb.onpaste = rfb.clipboardPasteFrom.bind(rfb);
      rfb.addEventListener('clipboard', (e: { detail: { text: string } }) => {
        if (!ref.current) return;
        const { text } = e.detail;
        const clipboardData = new DataTransfer();
        clipboardData.setData('text/plain', text);
        ref.current.dispatchEvent(
          new ClipboardEvent('copy', { clipboardData })
        );
      });
      cb.grab();
      setCb(cb);
    } else {
      errorHandler(new Error('unsupported_clipboard'));
    }

    setRfb(rfb);
  };

  useEffect(() => {
    if (!rfb) return;
    rfb.viewOnly = viewOnly;
    rfb.resizeSession = viewOnly;
    rfb.qualityLevel = quality;
    rfb.compressionLevel = compression;
  }, [compression, quality, rfb, viewOnly]);

  useEffect(() => {
    connect();
    return disconnect;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <div ref={ref} className="novnc-viewport absolute h-full w-full" />
    </>
  );
};

export default NoVnc;
