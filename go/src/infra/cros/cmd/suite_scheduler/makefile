# Copyright 2023 The Chromium OS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
STDOUTFILE=zout_log
STDERRFILE=zerr_log
BIN=suite_scheduler

FETCHEDSUSCHINI=configparser/fetched/suite_scheduler.ini
FETCHEDLABINI=configparser/fetched/lab_config.ini

all: run

run: build
	./$(BIN) >$(STDOUTFILE) 2>$(STDERRFILE)

build:
	go build -o $(BIN)	

clean:
	rm -rf $(STDERRFILE) $(STDOUTFILE) $(BIN) $(FETCHEDSUSCHINI) $(FETCHEDLABINI)

test:
	go test ./...

format:
	go fmt ./...