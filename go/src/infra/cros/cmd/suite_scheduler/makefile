# Copyright 2023 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
STDOUTFILE=zout_log
STDERRFILE=zerr_log
BIN=suite_scheduler

GENERATEDSUSCHCFG=configparser/generated/suite_scheduler.cfg
GENERATEDLABCFG=configparser/generated/lab_config.cfg
GENERATEDSUSCHINI=configparser/generated/suite_scheduler.ini
GENERATEDLABINI=configparser/generated/lab_config.ini

all: run

run: build
	./$(BIN) >$(STDOUTFILE) 2>$(STDERRFILE)

build:
	go build -o $(BIN)

clean:
	rm -rf $(STDERRFILE) $(STDOUTFILE) $(BIN) $(GENERATEDSUSCHCFG) $(GENERATEDLABCFG) $(GENERATEDSUSCHINI) $(GENERATEDLABINI)

test:
	go test ./...

format:
	go fmt ./...