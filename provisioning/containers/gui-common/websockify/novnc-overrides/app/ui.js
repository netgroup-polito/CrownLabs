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
    pathname = pathname
      .split('/')
      .filter((e) => e)
      .join('/');
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
    document.querySelector('#noVNC_connect_dlg .noVNC_logo').innerText = 'Connection failed';
    connAttempts = 0;
    oocp();
  }
};

/**
 *  STATS VIEW
 */
const link = document.createElement('link');
link.rel = 'stylesheet';
link.href = 'app/styles/metrics.css';
document.head.appendChild(link);

class Resources {
  constructor(cpu, mem, net, ip, timestamp, error) {
    if (cpu) this.cpu = Number(cpu);
    if (mem) this.mem = Number(mem);
    if (net) this.net = Number(net); //ms
    if (ip) this.ip = ip;
    if (timestamp) this.timestamp = timestamp;
    if (error) this.error = error;
  }

  static from(json) {
    if (json.cpu) json.cpu = Number(json.cpu) > 100 ? 100 : Number(json.cpu);
    if (json.net == undefined) json.net = 0;
    const resources = Object.assign(new Resources(), json);
    return resources;
  }
}

class ResourcesHistory {
  constructor(historyLen = 5) {
    this.lastResources = [];
    this.historyLen = historyLen;
  }

  /**
   * Save last resources
   */
  addResources(resources) {
    this.lastResources.push(resources);
    if (this.lastResources.length > this.historyLen) this.lastResources.shift();
  }

  getAvg(stat) {
    return this.lastResources.map((e) => e[stat]).reduce((a, b) => a + b, 0) / this.lastResources.length;
  }

  getCurrent(stat) {
    return this.lastResources.at(-1)[stat];
  }
}

/**
 * Register actions needed when DOM content is loaded:
 *  - Button drag and drop logic
 */
document.addEventListener('DOMContentLoaded', () => {
  /**
   * Button element
   */
  document.body.insertAdjacentHTML(
    'beforeend',
    `
  <div id="usage-btn" class="usage-btn">
    <span id="usage-btn-txt"></span>
    <img id="usage-btn-img" class="usage-img" src="app/images/stats-img.svg" alt="Stats"/>
    <div id="usage-img-transparent" class="usage-img-transparent"></div>
    <div id="usage-cpu" class="usage-cpu btn btn-grn" data-cont="CPU">
      <span id="usage-cpu-txt" class="usage-stat-txt">? %</span>
    </div>
    <div id="usage-mem" class="usage-mem btn btn-grn" data-cont="MEM">
      <span id="usage-mem-txt" class="usage-stat-txt">? %</span>
    </div>
    <div id="usage-net" class="usage-net btn btn-grn" data-cont="NET">
      <span id="usage-net-txt" class="usage-stat-txt">? ms</span>
    </div>
  </div>
  <div id="page-element" class="dim-screen"></div>`
  );

  const button = document.getElementById('usage-btn');
  const buttonNet = document.getElementById('usage-net');
  const buttonText = document.getElementById('usage-btn-txt');
  const overlay = document.getElementById('page-element');

  let resourcesHistory = new ResourcesHistory();

  const thresholds = {
    cpu: {
      warn: 80,
      crit: 90,
      el: document.getElementById('usage-cpu'),
      txt: document.getElementById('usage-cpu-txt'),
      format: (v) => `${v}%`,
    },
    mem: {
      warn: 85,
      crit: 90,
      el: document.getElementById('usage-mem'),
      txt: document.getElementById('usage-mem-txt'),
      format: (v) => `${v}%`,
    },
    net: {
      warn: 200,
      crit: 800,
      el: document.getElementById('usage-net'),
      txt: document.getElementById('usage-net-txt'),
      format: (v) => `${v}ms`,
    },
  };

  /**
   * DOM manupulation
   */
  let worstAvgColor = 'grn';
  const updateViewUsages = () => {
    button.classList.remove(`btn-${worstAvgColor}`);
    let worstColor = 'grn';
    worstAvgColor = 'grn';
    Object.keys(thresholds).forEach((stat) => {
      const { crit, warn, el, txt, format } = thresholds[stat];
      const value = resourcesHistory.getCurrent(stat);
      const avgValue = resourcesHistory.getAvg(stat);
      txt.innerHTML = format(value);
      let color = 'grn';
      if (value > crit) {
        color = 'red';
        worstColor = 'red';
      } else if (value > warn) {
        color = 'yel';
        if (worstColor === 'grn') worstColor = 'yel';
      }

      if (avgValue > crit) {
        worstAvgColor = 'red';
      } else if (avgValue > warn && worstAvgColor === 'grn') {
        worstAvgColor = 'yel';
      }

      el.className = `usage-${stat} btn btn-${color}`;
    });

    if (
      !((worstAvgColor === 'red' && worstColor === 'red') ||
      (worstAvgColor === 'yel' && (worstColor === 'yel' || worstColor === 'red')) ||
       worstAvgColor === 'grn')
    ) {
      worstAvgColor = worstColor;
    }
    button.classList.add(`btn-${worstAvgColor}`);
  };

  let isMoving = false;
  let initialX, initialY;

  const proto = window.location.protocol.replace('http','ws')
  const updatePeriod = 2; //seconds

  // WS metrics connection
  let metricsPath = `${proto}//${window.location.host}/${window.metricsTargetUrl}&updatePeriod=${updatePeriod}`;
  console.log('connceting usages ws ' + metricsPath);
  const conn = new WebSocket(metricsPath);
  conn.onclose = (evt) => {
    button.remove();
    Log.Error('Usages WS connection closed: ', evt);
  };
  conn.onerror = (err) => {
    button.remove();
    Log.Error('Usages WS connection error: ' + err);
  };
  conn.onmessage = (evt) => {
    try {
      const resourcesJson = JSON.parse(evt.data.toString());
      resourcesHistory.addResources(Resources.from(resourcesJson));
      updateViewUsages();
    } catch (error) {
      Log.Error('Error on JSON parsing from WS data: ', error);
      buttonText.innerHTML = 'WS Error';
    }
  };

  button.addEventListener('mousedown', (e) => {
    isMoving = true;
    initialX = e.clientX - button.offsetLeft;
    initialY = e.clientY - button.offsetTop;
    overlay.classList.add('active');
    button.classList.add('active');
  });

  document.addEventListener('mousemove', (e) => {
    if (isMoving) {
      let { clientX, clientY } = e;

      if (clientX - initialX < 0) clientX = initialX;
      if (clientX + button.offsetWidth - initialX > window.innerWidth)
        clientX = window.innerWidth - button.offsetWidth + initialX;
      if (clientY - initialY < -buttonNet.offsetTop) clientY = initialY - buttonNet.offsetTop;
      if (clientY > window.innerHeight - button.offsetHeight + initialY)
        clientY = window.innerHeight - button.offsetHeight + initialY;

      button.style.left = clientX - initialX + 'px';
      button.style.top = clientY - initialY + 'px';
    }
  });

  document.addEventListener('mouseup', () => {
    isMoving = false;
    overlay.classList.remove('active');
    button.classList.remove('active');
  });
});
