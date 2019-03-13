#!/bin/bash
docker run \
	-e INFRADOCK_NO_UPDATE=1  \
	-e LUCI_CONTEXT=/tmp/luci_context \
	-v /tmp/luci_context:/tmp/luci_context \
	gcr.io/chromium-container-registry/infra_dev_env \
	apack \
	pack \
	/source/infra/appengine/cr-buildbucket/default.apack