exploratory project for server-side in-memory session persistence

```
APACHE BENCH

chmod +x ab_test.sh


ulimit -n 65536

ab -n 1000000 -c 1000 -T 'application/json' -p ab_loadtest_payload.json http://localhost:8888/api/login

ab -n 1000000 -c 1000 -H "X-ACCESS-TOKEN: 6f44ceb34062006e9503bd577e7c7a8b187240413b2b.c83a99ea41fb8d76f0a8d35a0c7faa0f104cad0d45214a8425bfa6766c630aba" http://localhost:8888/api/resource


=================

NGINX

worker_processes auto;

events {
    use epoll; 
    worker_connections 4096; 
    multi_accept on; 
}

upstream app_chi_backend {
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
    server localhost:8083;
    server localhost:8084;
    keepalive 128;
}

server {
    listen 8888;

    gzip on;
    gzip_min_length 1000;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;


    location / {
        proxy_pass http://app_chi_backend; 
        proxy_http_version 1.1;

        proxy_buffering on;
        proxy_buffers 16 8k;
        proxy_busy_buffers_size 64k;

        proxy_connect_timeout 10s;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;

        proxy_next_upstream error timeout http_500 http_502 http_503 http_504;
        proxy_ignore_headers X-Accel-Expires Expires Cache-Control;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}




```


