# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# We copy the fleetcost binary into /app/fleetcost
# and serve from the /app directory.

FROM gcr.io/distroless/static:latest
WORKDIR /

COPY ./fleetcostserver /app/fleetcostserver
ENTRYPOINT ["/app/fleetcostserver"]
