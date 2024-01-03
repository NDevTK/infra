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
CODE_COVERAGE_MODULES := default code-coverage-frontend code-coverage-backend code-coverage-auxiliary-coverage-worker
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
	@echo "To run ALL tests: \`make test\`"
	@echo "To run tests in a sub directory: \`make test TEST_GLOB=services\`"
	@echo "To run a given test: \`make test TEST_GLOB=:*testUnrecognizedMasterUrl*\`"
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

install_deps:
	pip3 install -t lib -r requirements.txt

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



##### Unit tests ###
clean-expect_tests:  # $$ is to escape $ in makefile.
	ps auxef | grep ENV/bin/expect_tests | grep -v grep | \
	  awk '{print $$2}' | xargs kill -9 >/dev/null 2>&1 | echo

test: clean-expect_tests
ifeq ($(TEST_GLOB),"")
	@mkdir -p $(TMPDIR)/coverage
	$(INFRAPYTHON) ../../test.py test appengine/findit \
      --html-report $(TMPDIR)/coverage
else
	@echo "Note: code coverage disabled when running selected tests."
	@echo
	@mkdir -p $(TMPDIR)/coverage
	$(INFRAPYTHON) ../../test.py test appengine/findit/$(TEST_GLOB) \
      --html-report $(TMPDIR)/coverage --coveragerc appengine/findit/.coveragerc --no-coverage
endif

node_deps:
	npm install

wct: node_deps
	npx run-wct --base ui/ --dep third_party

wct_debug: node_deps
	npx run-wct --base ui/ --dep third_party --debug

##### TMP ###
TMPDIR := $(shell mkdir -p .tmp 2> /dev/null && realpath .tmp)
ifndef TMPDIR
TMPDIR :=  $(shell mktemp -d)
endif

.PHONY: _deploy_findit test update \
        help migrate deploy-findit-prod deploy-findit-staging \
        deploy-coverage-prod deploy-coverage-staging findit-predeploy \
        deploy-all-prod deploy-all-staging
