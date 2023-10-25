# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREDATOR_STAGING = predator-for-me-staging
PREDATOR_PROD = predator-for-me

MODULES= app.yaml \
	 backend-clusterfuzz.yaml \
	 backend-cracas.yaml \
	 backend-process.yaml \
	 backend-uma-sampling-profiler.yaml

default: help

help:
	@echo "Available commands:"
	@sed -n '/^[a-zA-Z0-9_.]*:/s/:.*//p' <Makefile

pydeps:
	python3 -m pip install --no-deps -t lib -r requirements.txt

test:
	../../test.py test appengine/predator

pytest:
	vpython3 test.py

deploy_staging:
	gcloud app deploy --no-cache --project ${PREDATOR_STAGING} ${MODULES}

deploy_prod:
	gcloud app deploy --project ${PREDATOR_PROD} ${MODULES}
