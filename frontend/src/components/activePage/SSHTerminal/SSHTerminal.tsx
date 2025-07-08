import React, { useContext, useEffect, useRef } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";
import { AuthContext } from '../../../contexts/AuthContext';

const { token } = useContext(AuthContext);

const hideScrollbarStyle: React.CSSProperties = {
  lineHeight: 1.2,
  width: "100%",
  height: "100vh",
  overflow: "hidden",
  scrollbarWidth: "none",
  msOverflowStyle: "none",
  backgroundColor: "black",
};

const injectGlobalStyle = () => {
  const style = document.createElement("style");
  style.innerHTML = `
    html, body {
      margin: 0;
      padding: 0;
      overflow: hidden;
      background-color: black;
    }
    ::-webkit-scrollbar {
      width: 0;
      background: transparent;
    }
  `;
  document.head.appendChild(style);
};

const SSHTerminal: React.FC = () => {
  const { namespace, name: nomeVM } = useParams<{ namespace: string; name: string }>();
  const [searchParams] = useSearchParams();
  const prettyName = searchParams.get("prettyName") ?? "";
  const containerRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);

  useEffect(() => {
    injectGlobalStyle();
    if (!containerRef.current) return;

  
    const term = new Terminal({
      cursorBlink: true,
      convertEol: true,
      scrollback: 10000,
      theme: { background: "#000000" },
    });
    term.open(containerRef.current);
    term.focus();
    termRef.current = term;

    const wsProtocol = "wss";
    const socket = new WebSocket(`${wsProtocol}://${window.location.host}/ws`);

    socket.onopen = () => {
      socket.send(
        JSON.stringify({
          namespace: namespace ?? "",
          vm: nomeVM ?? "",
          token: token ?? "",
        })
      );

      if (prettyName)
        term.writeln(`\x1b[1;36m📡 Connecting to: ${prettyName}\x1b[0m`);
      term.writeln("[✔] SSH connection success.\r\n");
    };

    socket.onmessage = (ev) => term.write(ev.data as string);
    socket.onerror = () => term.writeln("[✖] Connection error.\r\n");
    socket.onclose = () => term.writeln("[●] Connection closed.\r\n");

    term.onData((data) => {
      if (socket.readyState === WebSocket.OPEN) socket.send(data);
    });

    return () => {
      socket.close();
      term.dispose();
    };
  }, [namespace, nomeVM, prettyName]);

  return <div ref={containerRef} style={hideScrollbarStyle} />;
};

export default SSHTerminal;
