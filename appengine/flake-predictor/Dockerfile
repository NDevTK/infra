# This is a mix of
# https://cloud.google.com/appengine/docs/flexible/custom-runtimes/build
# and https://github.com/GoogleCloudPlatform/cloud-sdk-docker

FROM gcr.io/google_appengine/python

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -y -qq --no-install-recommends wget unzip && apt-get clean

WORKDIR /app

# Install the Google Cloud SDK.
ENV HOME /
ENV CLOUDSDK_PYTHON_SITEPACKAGES 1
RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && unzip google-cloud-sdk.zip && rm google-cloud-sdk.zip
RUN google-cloud-sdk/install.sh --usage-reporting=true --path-update=true --bash-completion=true --rc-path=/.bashrc --additional-components gsutil

# Disable updater check for the whole installation.
# Users won't be bugged with notifications to update to the latest version of
# gcloud.
RUN google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true

# Disable updater completely.
# Running `gcloud components update` doesn't really do anything in a union FS.
# Changes are lost on a subsequent run.
RUN sed -i -- 's/\"disable_updater\": false/\"disable_updater\": true/g' google-cloud-sdk/lib/googlecloudsdk/core/config.json

ENV PATH /app/google-cloud-sdk/bin:$PATH

RUN virtualenv /env -p python

# Set virtualenv environment variables. This is equivalent to running
# source /env/bin/activate
ENV VIRTUAL_ENV /env
ENV PATH /env/bin:$PATH
ADD requirements.txt /app/
RUN pip install -r requirements.txt
ADD . /app/
CMD gunicorn -b :$PORT dataflow:app
