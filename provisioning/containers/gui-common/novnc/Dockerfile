FROM nginx:1.19.5-alpine

ENV HTML_DATA=/usr/share/nginx/html

# Get noVNC & cleanup
RUN wget -qO- https://github.com/novnc/noVNC/archive/v1.2.0.tar.gz | \
    tar xz -C ${HTML_DATA} --strip 1 && \
    rm -rf ${HTML_DATA}/docs && \
    rm -rf ${HTML_DATA}/snap && \
    rm -rf ${HTML_DATA}/tests 

# Inject noVNC customizations (JS)
RUN echo "window.addEventListener('load', () => {\
let pathName = window.location.pathname.slice(1);\
if (pathName) UI.forceSetting('path', pathName+'/websockify');\
UI.forceSetting('resize', 'remote');\
UI.forceSetting('show_dot', true);\
UI.forceSetting('reconnect', true);\
UI.forceSetting('reconnect_delay', 100);\
UI.connect();\
})" >> ${HTML_DATA}/app/ui.js

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
