import React, { useEffect, useContext, useRef, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useXTerm } from 'react-xtermjs';
import { AuthContext } from '../../../contexts/AuthContext';
import './SSHTerminal.css';
import { FitAddon } from 'xterm-addon-fit';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import type { SupportedError } from '../../../errorHandling/utils';

type TerminalSize = {
  cols: number;
  rows: number;
};

const SSHTerminal: React.FC = () => {
  const { makeErrorCatcher } = useContext(ErrorContext);
  const genericErrorCatcher = useCallback(
    (error: SupportedError) => {
      makeErrorCatcher(ErrorTypes.GenericError)(error);
    },
    [makeErrorCatcher],
  );

  const { namespace = '', VMname: VmName = '', environment = '' } = useParams();
  const { ref, instance } = useXTerm();
  const { token } = useContext(AuthContext);

  const fitRef = useRef<FitAddon | null>(null);
  const resizeObs = useRef<ResizeObserver | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const keepAliveRef = useRef<number | null>(null);

  useEffect(() => {
    if (!instance) return;

    instance.options = {
      cursorBlink: true,
      scrollback: 10000,
      convertEol: true,
      theme: { background: '#000000' },
    };

    if (!fitRef.current) {
      fitRef.current = new FitAddon();
      instance.loadAddon(fitRef.current);
    }

    let lastSize: TerminalSize = { cols: instance.cols, rows: instance.rows };

    const fitAndNotify = () => {
      requestAnimationFrame(() => {
        fitRef.current?.fit();
        const { cols, rows } = instance;
        if (
          (cols !== lastSize.cols || rows !== lastSize.rows) &&
          wsRef.current?.readyState === WebSocket.OPEN
        ) {
          lastSize = { cols, rows };
          wsRef.current.send(JSON.stringify({ type: 'resize', cols, rows }));
        }
      });
    };

    instance.focus();
    fitAndNotify();
    window.addEventListener('resize', fitAndNotify);
    (document as Document & { fonts?: FontFaceSet }).fonts?.ready?.then(() =>
      fitAndNotify(),
    );

    if (ref.current && 'ResizeObserver' in window) {
      resizeObs.current = new ResizeObserver(fitAndNotify);
      resizeObs.current.observe(ref.current);
    }

    // --- WebSocket ---
    const PROT = location.protocol === 'https:' ? 'wss' : 'ws';
    const URL = location.host; // PRODUCTION
    // const URL = 'localhost:8090' // LOCAL - backend running on localhost
    // const URL = '950.staging.crownlabs.polito.it:80'; // STAGING

    const socketUrl = `${PROT}://${URL}/webssh`;
    const ws = new WebSocket(socketUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      // init message
      ws.send(
        JSON.stringify({
          namespace,
          vmName: VmName,
          token,
          InitialWidth: instance.cols,
          InitialHeight: instance.rows,
          Environment: environment,
        }),
      );

      instance.writeln('');
      instance.writeln('\x1b[1;36m📡 Connecting to VM... \x1b[0m');
      instance.writeln('\x1b[1;32m[✔] SSH connection success.\x1b[0m\r\n');

      fitAndNotify();

      // keepalive
      keepAliveRef.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }));
        }
      }, 30000 /* 30s */);
    };

    ws.onmessage = ev => {
      const obj = JSON.parse(ev.data);

      if (!obj.type) return;

      switch (obj.type) {
        case 'error': {
          instance.write(`\r\n\x1b[1;31m${obj.error}\x1b[0m\r\n`);
          ws.close();
          break;
        }
        case 'data': {
          instance.write(obj.data);
          break;
        }
        case 'pong':
        default: {
          // nothing to do
          break;
        }
      }
    };

    ws.onerror = () => {
      instance.writeln('\x1b[1;31m[✖] Connection error.\x1b[0m\r\n');
    };
    ws.onclose = () => {
      instance.writeln('\x1b[1;33m[●] Connection closed.\x1b[0m\r\n');
    };

    const disposeData = instance.onData(data => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }));
      }
    });

    return () => {
      disposeData.dispose();
      resizeObs.current?.disconnect();
      window.removeEventListener('resize', fitAndNotify);
      try {
        ws.close();
      } catch (error) {
        genericErrorCatcher(error as SupportedError);
      }
      wsRef.current = null;
      try {
        instance.dispose();
      } catch (error) {
        genericErrorCatcher(error as SupportedError);
      }
    };
  }, [
    instance,
    namespace,
    VmName,
    token,
    ref,
    environment,
    genericErrorCatcher,
  ]);

  return <div ref={ref} className="ssh-terminal" />;
};

export default SSHTerminal;
