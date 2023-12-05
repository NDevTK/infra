FROM ubuntu:16.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y puppet

RUN apt-get autoremove && \
    apt-get autoclean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/cache/man \
      /usr/share/{doc,groff,info,lintian,linda,man}
