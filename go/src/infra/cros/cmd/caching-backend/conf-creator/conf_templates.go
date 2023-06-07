// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

const keepalivedTemplate = `# This file is generated. DO NOT EDIT.

vrrp_script chk_caching_backend_health {
  script "curl http://localhost:8082/check_health"
  interval 3  # In second.
  weight 60
}

vrrp_instance CacheServer {
  state {{ .State }}
  interface {{ .Interface }}
  virtual_router_id 51
  priority {{ .Priority }}
  advert_int 1
  unicast_peer {
    {{ .UnicastPeer }}
  }
  authentication {
        auth_type PASS
        auth_pass PASSWORD
  }
  track_script {
    chk_caching_backend_health
  }
  virtual_ipaddress {
    {{ .VirtualIP }}
  }
}
`

const nginxTemplate = `# This file is generated. DO NOT EDIT.

user www-data;
worker_processes {{ if .WorkerCount }}{{ .WorkerCount }}{{ else }}auto{{ end }};
worker_rlimit_nofile 65535;

pid        /var/run/nginx.pid;
error_log  /var/log/nginx/error.log error;

{{ if .OtelTraceEndpoint }}
load_module /opt/opentelemetry-webserver-sdk/WebServerModule/Nginx/ngx_http_opentelemetry_module.so;
{{ end }}

events {
  accept_mutex on;
  accept_mutex_delay 500ms;
  worker_connections 1024;
}

http {
  include       /etc/nginx/mime.types;
  default_type  application/octet-stream;
  log_format main '$remote_addr - $remote_user [$time_iso8601] "$request" '
                  '$status $body_bytes_sent "$sent_http_content_length" '
                  '$request_time "$http_referer" '
                  '"$http_user_agent" "$http_x_forwarded_for" $upstream_cache_status';

  log_format main_json escape=json
  '{'
    '"access_time":"$time_iso8601",'
    '"bytes_sent":$body_bytes_sent,'
    '"content_length":$sent_http_content_length,'
    '"host":"$host",'
    '"method":"$request_method",'
    '"proxy_host":"$proxy_host",'
    '"referer":"$http_referer",'
    '"remote_addr":"$remote_addr",'
    '"remote_user":"$remote_user",'
    '"request":"$request",'
    '"request_time":$request_time,'
    '"status":$status,'
    '"uri":"$uri",'
    '"user_agent":"$http_user_agent",'
    '"upstream":"$upstream_addr",'
    '"upstream_cache_status":"$upstream_cache_status",'
    '"upstream_response_time":"$upstream_response_time",'
    '"swarming_task_id": "$http_x_swarming_task_id",'
    '"bbid": "$http_x_bbid",'
    '"x_forwarded_for":"$http_x_forwarded_for"'
  '}';

  proxy_cache_path  /var/cache/nginx levels=1:2 keys_zone=google-storage:80m
                    max_size={{ .CacheSizeInGB }}g inactive=720h;
  # gs_cache upstream definition.
  upstream gs_archive_servers {
    {{ if .UpstreamHost }}
    server {{ .UpstreamHost }} fail_timeout=10s;
    {{ range .Ports -}}
    server 127.0.0.1:{{.}} backup;
    {{ end -}}
    {{ else }}
    least_conn;
    {{ range .Ports -}}
    server 127.0.0.1:{{.}} fail_timeout=10s;
    {{ end -}}
    {{ end }}
  }
  server {
    listen *:8082;
    # TODO(guocb) Remove this after removing provision branch using gs_cache.
    listen *:8888;
    server_name           gs-cache;
    add_header            'Cache-Control' 'public, max-age=3153600';
    add_header            '{{ if .UpstreamHost }}X-Cache-Primary{{ else }}X-Cache-Secondary{{ end }}' '$upstream_cache_status';
    add_header            'X-CACHING-BACKEND' '$host';
    index  index.html index.htm index.php;
    access_log            /var/log/nginx/gs-cache.access.log main;
    access_log            /dev/stdout main_json;
    error_log             /var/log/nginx/gs-cache.error.log;
    error_log             /dev/stdout;


    # CQ build cache configuration.
    # The configuration is exactly same with the "location /" except
    # "proxy_cache_valid" which is much shorter than a release build.
    # A CQ build URL is like "/download/chromeos-image-archive/coral-cq/R92-13913.0.0-46943-8850024658050820208/...".
    location ~ ^/download/[^/]+/[^/]+-cq/ {
      # The two headers added below must be added in each location, instead of
      # in the "server" directive as it may not come as the request headers.
      # Instead, it may be set as variables in this configuration file, which
      # can only be seen after setting.
      add_header            'X-SWARMING-TASK-ID' '$http_x_swarming_task_id';
      add_header            'X-BBID' '$http_x_bbid';
      slice 30m;
      proxy_cache_lock on;
      proxy_cache_lock_age 900s;
      proxy_cache_lock_timeout 900s;
      proxy_cache_bypass $http_x_no_cache;
      expires max;
      proxy_pass            http://gs_archive_servers$uri$is_args$args;
      proxy_read_timeout    900;
      proxy_connect_timeout 90;
      proxy_redirect        off;
      proxy_http_version    1.1;
      proxy_set_header      Connection "";
      proxy_set_header      X-SWARMING-TASK-ID $http_x_swarming_task_id;
      proxy_set_header      X-BBID $http_x_bbid;
      proxy_set_header      X-Forwarded-Host {{ .VirtualIP }}:$server_port;
      proxy_set_header      X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_cache           google-storage;
      proxy_cache_valid     200 206 48h;
      proxy_cache_key       $request_method$uri$is_args$args$slice_range;
      proxy_set_header      Range $slice_range;
      proxy_force_ranges    on;
    }
    location ~ ^/[^/]+/[^/]+/[^/]+-cq/ {
      add_header            'X-SWARMING-TASK-ID' '$http_x_swarming_task_id';
      add_header            'X-BBID' '$http_x_bbid';
      proxy_cache_lock on;
      proxy_cache_lock_age 3600s;
      proxy_cache_lock_timeout 3600s;
      proxy_cache_bypass $http_x_no_cache;
      expires max;
      proxy_pass            http://gs_archive_servers$uri$is_args$args;
      proxy_read_timeout    3600;
      proxy_connect_timeout 90;
      proxy_redirect        off;
      proxy_http_version    1.1;
      proxy_set_header      Connection "";
      proxy_set_header      X-SWARMING-TASK-ID $http_x_swarming_task_id;
      proxy_set_header      X-BBID $http_x_bbid;
      proxy_set_header      X-Forwarded-Host {{ .VirtualIP }}:$server_port;
      proxy_set_header      X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_cache           google-storage;
      proxy_cache_valid     200 48h;
      proxy_cache_key       $request_method$uri$is_args$args;
    }

    location / {
      add_header            'X-SWARMING-TASK-ID' '$http_x_swarming_task_id';
      add_header            'X-BBID' '$http_x_bbid';
      proxy_cache_lock on;
      proxy_cache_lock_age 3600s;
      proxy_cache_lock_timeout 3600s;
      proxy_cache_bypass $http_x_no_cache;
      expires max;
      proxy_pass            http://gs_archive_servers$uri$is_args$args;
      proxy_read_timeout    3600;
      proxy_connect_timeout 90;
      proxy_redirect        off;
      proxy_http_version    1.1;
      proxy_set_header      Connection "";
      proxy_set_header      X-SWARMING-TASK-ID $http_x_swarming_task_id;
      proxy_set_header      X-BBID $http_x_bbid;
      proxy_set_header      X-Forwarded-Host {{ .VirtualIP }}:$server_port;
      proxy_set_header      X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_cache           google-storage;
      proxy_cache_valid     200 720h;
      proxy_cache_key       $request_method$uri$is_args$args;
    }
    location ~ ^/download/ {
        add_header            'X-SWARMING-TASK-ID' '$http_x_swarming_task_id';
        add_header            'X-BBID' '$http_x_bbid';
        slice 30m;
        proxy_cache_lock on;
        proxy_cache_lock_age 900s;
        proxy_cache_lock_timeout 900s;
        proxy_cache_bypass $http_x_no_cache;
        expires max;
        proxy_pass            http://gs_archive_servers$uri$is_args$args;
        proxy_read_timeout    900;
        proxy_connect_timeout 90;
        proxy_redirect        off;
        proxy_http_version    1.1;
        proxy_set_header      Connection "";
        proxy_set_header      X-SWARMING-TASK-ID $http_x_swarming_task_id;
        proxy_set_header      X-BBID $http_x_bbid;
        proxy_set_header      X-Forwarded-Host {{ .VirtualIP }}:$server_port;
        proxy_set_header      X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_cache           google-storage;
        proxy_cache_valid     200 206 720h;
        proxy_cache_key       $request_method$uri$is_args$args$slice_range;
        proxy_set_header      Range $slice_range;
        proxy_force_ranges    on;
      }

    # b/281868022: A location dedicated for AU tests.
    location ~ ^/swarming/(?<swarming>[^/]+)/bbid/(?<bbid>[^/]+)/(?<end>.*)$ {
        set $http_x_swarming_task_id "$swarming";
        set $http_x_bbid "$bbid";
        rewrite .* "/$end" last;
    }

    # Rewrite rules converting devserver client requests to gs_cache.
    location @gs_cache {
      if ($arg_gs_bucket = "") {
        set $arg_gs_bucket "chromeos-image-archive";
      }
      # The ending '?' erase any query string from the incoming request.
      rewrite "^/static/(tast/cros/.+)" "/download/chromiumos-test-assets-public/$1?" last;
      rewrite "^/static/(tast/.+)" "/download/chromeos-test-assets-private/$1?" last;
      rewrite "^/static/([^/]+-channel/.+)$" "/download/chromeos-releases/$1?" last;
      rewrite "^/static/([^/]+/[^/]+)/(autotest/packages)/(.*)" "/extract/$arg_gs_bucket/$1/autotest_packages.tar?file=$2/$3?" last;
      rewrite "^/static/([^/]+/[^/]+/chromiumos_test_image)\.bin$" "/extract/$arg_gs_bucket/$1.tar.xz?file=chromiumos_test_image.bin?" last;
      rewrite "^/static/([^/]+/[^/]+/recovery_image)\.bin$" "/extract/$arg_gs_bucket/$1.tar.xz?file=recovery_image.bin?" last;
      rewrite "^/static/(.+)$" "/download/$arg_gs_bucket/$1?" last;
    }
    # Some legacy RPCs in order to be backward compatible with devserver.
    location /check_health {
      default_type application/json;
      return 200 '{"disk_total_bytes_per_second": 0, "network_total_bytes_per_second": 0, "network_sent_bytes_per_second": 0, "apache_client_count": 0, "disk_write_bytes_per_second": 0, "cpu_percent": 0, "disk_read_bytes_per_second": 0, "gsutil_count": 0, "network_recv_bytes_per_second": 0, "free_disk": 5678, "au_process_count": 0, "staging_thread_count": 0, "telemetry_test_count": 0}';
    }
    location /stage {
      return 200 'Success';
    }
    location /is_staged {
      return 200 'True';
    }
    location = /download/chromeos-image-archive {
      return 400;
    }
    location = /static {
      alias /var/www/nginx_static;
      autoindex on;
    }
    location /static/ {
      alias /var/www/nginx_static/;
      try_files $uri @gs_cache;
    }
    location /list_image_dir {
      return 200 'The /list_image_dir RPC is not supported by GS Cache. Usage is discouraged.';
    }
  }
{{ if .OtelTraceEndpoint }}
    NginxModuleEnabled ON;
    NginxModuleOtelSpanExporter otlp;
    NginxModuleOtelExporterEndpoint {{ .OtelTraceEndpoint }};
    NginxModuleServiceName CachingBackendNginx;
    NginxModuleServiceNamespace CachingBackendNginx;
    NginxModuleServiceInstanceId CachingBackendNginxId;
    NginxModuleResolveBackends ON;
{{ end }}
}
`

// Non operational config templates.

const noOpKeepalivedTemplate = `# This file is generated. DO NOT EDIT.
# This file is intentionally empty.
`

const noOpNginxTemplate = `# This file is generated. DO NOT EDIT.

events {}
`
