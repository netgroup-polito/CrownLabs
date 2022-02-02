FROM nginx:1.21.6-alpine

ENV HTML_DATA=/usr/share/nginx/html

# Get noVNC & cleanup
RUN wget -qO- https://github.com/novnc/noVNC/archive/v1.3.0.tar.gz | \
    tar xz -C ${HTML_DATA} --strip 1 && \
    rm -rf ${HTML_DATA}/docs && \
    rm -rf ${HTML_DATA}/snap && \
    rm -rf ${HTML_DATA}/tests 

# Inject noVNC customizations (JS)
RUN echo "const ogcv = getConfigVar;\
getConfigVar = (a,b) => {\
    if (a === 'autoconnect') return true;\
    return ogcv(a,b);\
}" >> ${HTML_DATA}/app/webutil.js

RUN echo "window.onhashchange=()=>{ \
const { hash } = window.location; \
UI.forceSetting('resize', hash.includes('noresize') ? 'scale' : 'remote');\
UI.forceSetting('view_only', hash.includes('readonly')); \
UI.updateViewOnly(); \
UI.applyResizeMode(); \
}; \
const ois = UI.initSettings;\
UI.initSettings = () => {\
ois();\
let {pathname} = window.location;\
pathname = pathname.split('/').filter(e=>e).join('/');\
if (pathname) UI.forceSetting('path', pathname+'/vnc');\
UI.forceSetting('show_dot', true);\
UI.forceSetting('reconnect', true);\
UI.forceSetting('reconnect_delay', Math.floor(Math.random()*2000));\
window.onhashchange();\
}" >> ${HTML_DATA}/app/ui.js

RUN echo #noVNC_container
RUN echo "#noVNC_container {border-bottom-right-radius: 0 !important;}" \
    >> ${HTML_DATA}/app/styles/base.css

# Copy config template and entrypoint
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY entrypoint.sh /entrypoint.sh

# Set default build arguments
ARG WEBSOCKIFY_HOST=localhost
ARG WEBSOCKIFY_PORT=8888
ARG HTTP_PORT=8080

ARG UID=1010
ARG USER=crownlabs

# Create default env variables
ENV WEBSOCKIFY_HOST=${WEBSOCKIFY_HOST}\
    WEBSOCKIFY_PORT=${WEBSOCKIFY_PORT}\
    HTTP_PORT=${HTTP_PORT}\
    HIDE_NOVNC_BAR=false\
    USER=${USER}

# Add non-root user and ensure authorizations
RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER} && \
    chown -R ${USER}:${USER} $HTML_DATA && \
    chown -R ${USER}:${USER} /var/cache/nginx && \
    chown -R ${USER}:${USER} /var/log/nginx && \
    chown -R ${USER}:${USER} /etc/nginx/nginx.conf && \
    touch /var/run/nginx.pid && \
    chown ${USER}:${USER} /var/run/nginx.pid && \
    chmod a+x /entrypoint.sh

EXPOSE ${HTTP_PORT}

# Enforce non-root user
USER ${USER}

ENTRYPOINT [ "/entrypoint.sh" ]
