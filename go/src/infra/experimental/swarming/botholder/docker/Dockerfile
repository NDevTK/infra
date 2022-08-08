FROM ubuntu:20.04

WORKDIR /opt/swarming

COPY setup.sh shutdown.sh botholder ./
RUN ./setup.sh

USER swarming:swarming
ENTRYPOINT ["./botholder"]
