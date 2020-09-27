import React, { useState, useEffect } from 'react';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Body from './Body';
import Header from './Header';
import Footer from './Footer';

const lightThemeConfig = {
  palette: {
    primary: { main: '#013378' },
    secondary: { main: '#FF7C11' }
  }
};
lightThemeConfig.palette.type = 'light';
// don't know why but spread operator doesn't work
const darkThemeConfig = JSON.parse(JSON.stringify(lightThemeConfig));
darkThemeConfig.palette.type = 'dark';
const lightTheme = createMuiTheme(lightThemeConfig);
const darkTheme = createMuiTheme(darkThemeConfig);

const Themer = props => {
  const {
    logout,
    name,
    isStudentView,
    adminGroups,
    registryName,
    imageList,
    templateLabsAdmin,
    instanceLabsAdmin,
    templateLabs,
    instanceLabs,
    changeAdminView,
    createTemplate,
    startCRDinstance,
    deleteCRDtemplate,
    connect,
    connectAdmin,
    stopCRDinstance,
    stopCRDinstanceAdmin
  } = props;

  const [isLightTheme, setIsLightTheme] = useState(() => {
    const prevIsLightTheme = JSON.parse(localStorage.getItem('isLightTheme'));
    if (prevIsLightTheme === null) return false;
    return prevIsLightTheme;
  });
  useEffect(() => {
    localStorage.setItem('isLightTheme', JSON.stringify(isLightTheme));
  }, [isLightTheme]);

  return (
    <ThemeProvider theme={isLightTheme ? lightTheme : darkTheme}>
      <CssBaseline />
      <div style={{ height: '100%' }}>
        <Header
          logged
          logout={logout}
          name={name}
          isStudentView={isStudentView}
          switchAdminView={changeAdminView}
          isLightTheme={isLightTheme}
          setIsLightTheme={setIsLightTheme}
        />
        <Body
          registryName={registryName}
          retrieveImageList={imageList}
          adminGroups={adminGroups}
          templateLabsAdmin={templateLabsAdmin}
          instanceLabsAdmin={instanceLabsAdmin}
          templateLabs={templateLabs}
          createNewTemplate={createTemplate}
          instanceLabs={instanceLabs}
          start={startCRDinstance}
          deleteLabTemplate={deleteCRDtemplate}
          connect={connect}
          connectAdmin={connectAdmin}
          stop={stopCRDinstance}
          stopAdmin={stopCRDinstanceAdmin}
          isStudentView={isStudentView}
        />
        <Footer />
      </div>
    </ThemeProvider>
  );
};

export default Themer;
