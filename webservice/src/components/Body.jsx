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
  const {
    isStudentView,
    registryName,
    retriveImageList,
    adminGroups,
    templateLabsAdmin,
    instanceLabsAdmin,
    connectAdmin,
    showStatus,
    createNewTemplate,
    start,
    stopAdmin,
    deleteLabTemplate,
    templateLabs,
    instanceLabs,
    connect,
    stop
  } = props;

  return (
    <ThemeProvider theme={theme === 'light' ? lightTheme : darkTheme}>
      <div
        style={{
          // the height of the container is viewport heigh - header height(70) - footer height(70)
          height: 'calc(100vh - 134px)',
          overflow: 'auto'
        }}
      >
        {isStudentView ? (
          <StudentView
            templateLabs={templateLabs}
            instanceLabs={instanceLabs}
            start={start}
            connect={connect}
            stop={stop}
            showStatus={showStatus}
          />
        ) : (
          <ProfessorView
            registryName={registryName}
            imageList={retriveImageList}
            adminGroups={adminGroups}
            templateLabs={templateLabsAdmin}
            instanceLabs={instanceLabsAdmin}
            connect={connectAdmin}
            showStatus={showStatus}
            createNewTemplate={createNewTemplate}
            start={start}
            stop={stopAdmin}
            deleteLabTemplate={deleteLabTemplate}
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
