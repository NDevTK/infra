# Copyright 2023 The LUCI Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Use ubuntu as the app depends on `tail` command util.
FROM ubuntu:22.04

WORKDIR /
# The binary is built on the host. Copy it into the image.
COPY nginx-access-log-metrics .

USER nobody
ENTRYPOINT ["./nginx-access-log-metrics"]
