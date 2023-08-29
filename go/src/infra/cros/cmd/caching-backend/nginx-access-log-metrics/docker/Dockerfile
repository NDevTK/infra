# Copyright 2023 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Use ubuntu as the app depends on `tail` command util.
FROM ubuntu:22.04
# This is necessary because we need cacerts to talk to BigQuery.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        && rm -rf /var/lib/apt/lists/*

WORKDIR /
# The binary is built on the host. Copy it into the image.
COPY nginx-access-log-metrics .

# The binary need to read logs mounted from host which is owned by root.
USER root
ENTRYPOINT ["./nginx-access-log-metrics"]
