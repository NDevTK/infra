# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

INFRAPYTHON := "../../ENV/bin/python"
GAE_PY := "./gae.py"
APPID := "findit-for-me"
STAGING_APPID := "findit-for-me-staging"
# Should come with infra repository, if this step fails run
# ...$ gclient sync
PROTOC := "../../cipd/protoc"
PROTO_PATH ?= "model/proto/"
CODE_COVERAGE_MODULES := code-coverage-frontend
LOCAL_VERSION := `$(INFRAPYTHON) $(GAE_PY) version`
LOCAL_HASH := `git rev-parse --short HEAD`
# If there are multiple active versions just use the latest one.
REMOTE_FINDIT_VERSION := `$(INFRAPYTHON) $(GAE_PY) active -A $(APPID) | grep "default:" | awk '{print $$2}'`
REMOTE_FINDIT_HASH := `$(INFRAPYTHON) $(GAE_PY) active -A $(APPID) | grep "default:" | awk '{print $$2}' | cut -d '-' -f2`
REMOTE_COVERAGE_VERSION := `$(INFRAPYTHON) $(GAE_PY) active -b -A $(APPID) | grep "code-coverage-backend:" | awk '{print $$2}'`
REMOTE_COVERAGE_HASH := `$(INFRAPYTHON) $(GAE_PY) active -b -A $(APPID) | grep "code-coverage-backend:" | awk '{print $$2}' | cut -d '-' -f2`
TEST_GLOB ?= ""

help:
	@echo "Code Coverage:"
	@echo "To deploy a new version of Code Coverage, check out the appropriate" \
	      "revision and run \`gclient sync\` then run" \
	      "\`make deploy-coverage-prod\` for production or \`make deploy-coverage-staging\`"\
	      "for staging."
	@echo
	@echo "Run tests"
	@echo "To run ALL tests: \`make pytest\`"
	@echo "To run tests in a sub directory: \`make pytest TEST_GLOB=services\`"
	@echo "To run a given test: \`make pytest TEST_GLOB=:*testUnrecognizedMasterUrl*\`"
	@echo
	@echo "To format code run \`make format\`"
	@echo
	@echo "To run all modules locally use \`make run\`"


##### App Engine ###
findit-predeploy:
	@echo "Uncommitted local changes (Commit, revert or delete):"
	@git status --porcelain
	@echo
	@echo "Changes in local branch. (Upload/land/pull/switch to master to avoid deploying tainted version)"
	@git log origin/master..HEAD --pretty=oneline .
	@echo
	@echo "Deployable Changes. (git pull and client sync if not current)"
	@git log $(REMOTE_FINDIT_HASH)..HEAD --pretty=oneline .

deploy-coverage-prod: pytest
	$(GAE_PY) upload -x -A $(APPID) $(CODE_COVERAGE_MODULES);
	@echo "New version is https://$(LOCAL_VERSION)-dot-code-coverage-frontend-"\
	"dot-$(APPID).appspot.com/p/chromium/coverage"

deploy-coverage-staging: pytest
	$(GAE_PY) upload -x -A $(STAGING_APPID) $(CODE_COVERAGE_MODULES);
	@echo "New version is https://$(LOCAL_VERSION)-dot-code-coverage-frontend-"\
	"dot-$(STAGING_APPID).appspot.com/p/chromium/coverage"


generate_requirements_txt:
	vpython3 -m pip freeze > requirements.txt
	printf "# This file is generated from the .vpython3 spec file.\n# Use \`make generate_requirements_txt\` to update.\n$$(cat requirements.txt)" > requirements.txt

pytest:
	vpython3 test.py $(TEST_GLOB)

##### Development ###
run:
	$(INFRAPYTHON) $(GAE_PY) devserver

YAPF := $(shell command -v yapf 2> /dev/null)
format:
ifndef YAPF
	@echo "Please install yapf first: https://github.com/google/yapf." && false
endif
	git cl format --full --python

proto:
	@$(PROTOC) --proto_path=$(PROTO_PATH) \
		--python_out=$(PROTO_PATH)gen \
		$(PROTO_PATH)*.proto
