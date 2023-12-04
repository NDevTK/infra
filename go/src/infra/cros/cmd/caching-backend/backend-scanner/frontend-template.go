// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

// keepalivedTemplate is for the configuration of keepalived, the caching
// service frontend, i.e. the load balancer.
const keepalivedTempalte = `
# This file is generated. DO NOT EDIT!

global_defs {
	process_names
	lvs_flush

	# The below "lvs_timeouts" depends on "lvs_sync_daemon".
	lvs_sync_daemon bond0 VI_Cache

	# During /extract or /decompress of our caching RPC, there may be long time
	# that's no data transported back to the client. Thus increase the TCP_CHECK
	# Timeout so LVS won't reset the connections.
	lvs_timeouts tcp 3600
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

	# When health checking failed, instead of removing the real server from
	# the LVS table (which resets the connection), we just set its weight to 0 in
	# order to make LVS keep forwarding packets to the real server. This is very
	# important to for our backend Nginx to drain the connections.
	inhibit_on_failure

	{{ range .RealServers -}}
	real_server {{ .IP }} {{ $.ServicePort }}{
		weight {{if .Terminating}}0{{else}}1{{end}}
		TCP_CHECK {
		  connect_timeout 5
		  connect_port {{ $.ServicePort }}
		}
	}
	{{ end -}}
}
`
