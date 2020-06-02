import React, { useState } from 'react';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import ProfessorView from '../views/ProfessorView';
import StudentView from '../views/StudentView';

const ChatSupport = () => {
  !(function () {
    const t = (window.driftt = window.drift = window.driftt || []);
    if (!t.init) {
      if (t.invoked)
        return void (
          window.console &&
          console.error &&
          console.error('Drift snippet included twice.')
        );
      (t.invoked = !0),
        (t.methods = [
          'identify',
          'config',
          'track',
          'reset',
          'debug',
          'show',
          'ping',
          'page',
          'hide',
          'off',
          'on'
        ]),
        (t.factory = function (e) {
          return function () {
            const n = Array.prototype.slice.call(arguments);
            return n.unshift(e), t.push(n), t;
          };
        }),
        t.methods.forEach(function (e) {
          t[e] = t.factory(e);
        }),
        (t.load = function (t) {
          const e = 3e5;
          const n = Math.ceil(new Date() / e) * e;
          const o = document.createElement('script');
          (o.type = 'text/javascript'),
            (o.async = !0),
            (o.crossorigin = 'anonymous'),
            (o.src = `https://js.driftt.com/include/${n}/${t}.js`);
          const i = document.getElementsByTagName('script')[0];
          i.parentNode.insertBefore(o, i);
        });
    }
  })();
  drift.SNIPPET_VERSION = '0.3.1';
  drift.load('SECRET');
};

export default function Body(props) {
  ChatSupport();
  const lightTheme = React.useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: 'light'
        }
      }),
    []
  );

  const darkTheme = React.useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: 'dark'
        }
      }),
    []
  );

  const [theme, setTheme] = useState('light');
  const toggleTheme = () => {
    if (theme === 'light') {
      setTheme('dark');
      document.getElementById('body').style.background = '#303030';
      document.getElementById('toolbar').style.background = '#424242';
      document.getElementById('footer').style.background = '#424242';
    } else {
      setTheme('light');
      document.getElementById('body').style.background = '#FAFAFA';
      document.getElementById('toolbar').style.background = '#032364';
      document.getElementById('footer').style.background = '#032364';
    }
  };

  return (
    <ThemeProvider
      theme={theme === 'light' ? lightTheme : darkTheme}
      // the height of the container is viewport heigh - header height(70) - footer height(70)
    >
      <div
        style={{
          height: 'calc(100vh - 134px)',
          overflow: 'auto'
        }}
      >
        {!props.adminHidden ? (
          <ProfessorView
            registryName={props.registryName}
            imageList={props.retriveImageList}
            adminGroups={props.adminGroups}
            templateLabs={props.templateLabsAdmin}
            instanceLabs={props.instanceLabsAdmin}
            events={props.events}
            funcTemplate={props.funcTemplate}
            funcInstance={props.funcInstance}
            connect={props.connectAdmin}
            showStatus={props.showStatus}
            hidden={props.hidden}
            funcNewTemplate={props.funcNewTemplate}
            start={props.start}
            stop={props.stopAdmin}
            delete={props.delete}
          />
        ) : (
          <StudentView
            templateLabs={props.templateLabs}
            instanceLabs={props.instanceLabs}
            funcTemplate={props.funcTemplate}
            funcInstance={props.funcInstance}
            start={props.start}
            connect={props.connect}
            stop={props.stop}
            events={props.events}
            showStatus={props.showStatus}
            hidden={props.hidden}
          />
        )}
      </div>
      <IconButton
        id="themeSwitch"
        style={{ position: 'absolute' }}
        onClick={toggleTheme}
      />
    </ThemeProvider>
  );
}
