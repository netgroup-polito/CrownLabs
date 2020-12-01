FROM nginx:1.19.5-alpine

ENV HTML_DATA=/usr/share/nginx/html

# Get noVNC & cleanup
RUN wget -qO- https://github.com/novnc/noVNC/archive/v1.2.0.tar.gz | \
    tar xz -C $HTML_DATA --strip 1 && \
    rm -rf $HTML_DATA/vendor/browser-es-module-loader/dist/babel-worker.js && \
    rm -rf $HTML_DATA/docs && \
    rm -rf $HTML_DATA/snap && \
    rm -rf $HTML_DATA/tests 

# Inject noVNC customizations (JS)
RUN echo "window.addEventListener('load', () => {\
UI.forceSetting('resize', 'remote');\
UI.forceSetting('show_dot', true);\
UI.forceSetting('reconnect', true);\
UI.forceSetting('reconnect_delay', 100);\
UI.connect();\
})" >> $HTML_DATA/app/ui.js

# Inject noVNC customizations (CSS: hide toolbar)
RUN echo "#noVNC_control_bar_anchor {display:none !important;}" >> \
$HTML_DATA/app/styles/base.css

COPY nginx.conf.template /etc/nginx/nginx.conf.template

ARG WEBSOCKIFY_HOST=localhost
ARG WEBSOCKIFY_PORT=8888
ARG HTTP_PORT=8080

ARG UID=1010
ARG USER=crownlabs

ENV WEBSOCKIFY_HOST=${WEBSOCKIFY_HOST}\
    WEBSOCKIFY_PORT=${WEBSOCKIFY_PORT}\
    HTTP_PORT=${HTTP_PORT}\
    USER=${USER}

RUN adduser -D -H -u ${UID} -s /sbin/nologin ${USER}

#Finalize config
RUN envsubst '${WEBSOCKIFY_HOST} ${WEBSOCKIFY_PORT} ${HTTP_PORT}' \
    < /etc/nginx/nginx.conf.template \
    > /etc/nginx/nginx.conf

RUN chown -R ${USER}:${USER} $HTML_DATA && \
    chown -R ${USER}:${USER} /var/cache/nginx && \
    chown -R ${USER}:${USER} /var/log/nginx && \
    chown -R ${USER}:${USER} /etc/nginx/nginx.conf && \
    touch /var/run/nginx.pid && \
    chown -R ${USER}:${USER} /var/run/nginx.pid 

EXPOSE ${HTTP_PORT}

USER ${USER}

# Start nginx (not in daemon mode for Docker use)
CMD nginx -g "daemon off;"
