events {
  worker_connections 256;
}

daemon off;

http {
  server {
    include /etc/nginx/mime.types;

    listen ${LISTEN_PORT};
    
    # WebSocket proxy for /websockify
    location /${BASE_PATH}websockify {
      proxy_pass http://localhost:5900;
      proxy_http_version 1.1;
      proxy_set_header Upgrade         $http_upgrade;
      proxy_set_header Connection      "Upgrade";
      proxy_set_header Host            $http_host;
      proxy_read_timeout 86400;
    }

    # Proxy for /filebrowser
    location /${BASE_PATH}files {
      proxy_pass http://localhost:8081;
      # proxy_http_version 1.1;
      # proxy_set_header Upgrade         $http_upgrade;
      # proxy_set_header Connection      "Upgrade";
      # proxy_set_header Host            $http_host;
    }
    
    location /${BASE_PATH} {
      alias /usr/share/novnc/;
    }
  }
}
