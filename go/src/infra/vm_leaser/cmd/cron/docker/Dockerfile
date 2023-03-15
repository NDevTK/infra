# Copyright 2023 The ChromiumOS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

FROM gcr.io/distroless/static:latest

COPY cron .

USER nobody

ENTRYPOINT ["./cron"]
