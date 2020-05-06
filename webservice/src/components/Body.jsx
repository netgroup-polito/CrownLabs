import React, { useState } from 'react';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import ProfessorView from '../views/ProfessorView';
import StudentView from '../views/StudentView';

export default function Body(props) {
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
            templateLabs={props.templateLabs}
            instanceLabs={props.instanceLabs}
            events={props.events}
            funcTemplate={props.funcTemplate}
            funcInstance={props.funcInstance}
            connect={props.connect}
            showStatus={props.showStatus}
            hidden={props.hidden}
            createTemplate={props.funcNewTemplate}
            start={props.start}
            stop={props.stop}
            // createLab={this.apiManager.createCRDinstance(this.MyName,this.MyNamespace)}
            // deleteLab={this.apiManager.deleteCRDinstance(this.MyName)}
            // enableOdisable={this.apiManager.setCRDinstanceStatus(CRDinstanceStatus)}
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
