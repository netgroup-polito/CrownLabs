window.onhashchange = () => {
  const { hash } = window.location;
  UI.forceSetting('resize', hash.includes('noresize') ? 'scale' : 'remote');
  UI.forceSetting('view_only', hash.includes('readonly'));
  UI.updateViewOnly();
  UI.applyResizeMode();
};
const ois = UI.initSettings;
UI.initSettings = () => {
  ois();
  if (window.websockifyTargetUrl) {
    UI.forceSetting('path', window.websockifyTargetUrl);
  } else {
    let { pathname } = window.location;
    pathname = pathname.split('/').filter(e => e).join('/');
    if (pathname) UI.forceSetting('path', pathname + '/vnc');
  }
  UI.forceSetting('show_dot', true);
  UI.forceSetting('reconnect', true);
  UI.forceSetting('reconnect_delay', Math.floor(Math.random() * 2000));
  window.onhashchange();
}
const oocp = UI.openConnectPanel;
let connAttempts = 0;
UI.openConnectPanel = () => {
  connAttempts++;
  if (connAttempts < 5) {
    setTimeout(UI.connect, 100);
  } else {
    document.querySelector('#noVNC_connect_dlg .noVNC_logo').innerText = 'Connection failed'
    connAttempts = 0;
    oocp();
  }
}
