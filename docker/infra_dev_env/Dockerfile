# Use an official Python runtime as a parent image
FROM google/cloud-sdk

# Install depot_tools
RUN git clone https://chromium.googlesource.com/chromium/tools/depot_tools/ /depot_tools
ENV PATH="/source/depot_tools:${PATH}"

# # Set the working directory to /app

# Fetch infra repo.
WORKDIR /source
RUN /depot_tools/fetch infra


# Install go toolchain.
# Includes node and npm.
RUN /source/infra/go/env.py

# Install bower
RUN /source/infra/go/env.py npm install -g bower@1.8.8


# Run commands within go env.
COPY entrypoint.sh /container-bootstrap/entrypoint.sh
ENTRYPOINT ["/container-bootstrap/entrypoint.sh"]


# Hacks
ADD apack /source/infra/go/bin
ADD buildbucket-default.apack /source/infra/appengine/cr-buildbucket/default.apack
