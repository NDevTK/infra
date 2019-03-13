#!/usr/bin/env bash
set -e

if [ -z $INFRADOCK_NO_UPDATE ]; then
	update_depot_tools
	SYNC_ARGS=
	if [ $INFRADOCK_INFRA_REV ]; then
		SYNC_ARGS="-r infra@$INFRADOCK_INFRA_REV"
	fi
	cd /source && gclient sync $SYNC_ARGS
	/source/infra/go/env.py echo done
	echo --- ENV UPDATED------------------------------------------------------------
	echo
fi

/source/infra/go/env.py $@
