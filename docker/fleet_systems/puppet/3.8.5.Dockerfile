FROM ubuntu:16.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y puppet=3.8.5-2ubuntu0.1

RUN apt-get autoremove && \
    apt-get autoclean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/cache/man \
      /usr/share/{doc,groff,info,lintian,linda,man}
