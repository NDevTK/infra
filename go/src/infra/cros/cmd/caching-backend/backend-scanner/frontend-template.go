// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

// keepalivedTempalte is for the configuration of keepalived, the caching
// service frontend, i.e. the load balancer.
const keepalivedTempalte = `
# This file is generated. DO NOT EDIT!

global_defs {
	process_names
	lvs_flush
}

vrrp_track_file goto_backup {
	file /goto_backup
	weight 0
	init_file 0
}

vrrp_instance VI_Cache {
	state BACKUP
	interface {{ .Interface }}
	virtual_router_id 52
	priority 150
	advert_int 1
	virtual_ipaddress {
		{{ .ServiceIP }}
	}
	track_file {
		goto_backup
	}
}

virtual_server {{ .ServiceIP }} {{ .ServicePort }}{
	delay_loop 30
	lb_algo {{ .LBAlgo }}
	lb_kind DR
	persistence_timeout 0 # to force the RR
	protocol TCP

	{{ range .RealServers -}}
	real_server {{ . }} {{ $.ServicePort }}{
		weight 1
		TCP_CHECK {
		  connect_timeout 5
		  connect_port {{ $.ServicePort }}
		}
	}
	{{ end -}}
}
`
