FROM ubuntu:trusty

ENV DEBIAN_FRONTEND noninteractive

RUN /usr/bin/apt-get update
RUN /usr/bin/apt-get -y dist-upgrade
RUN /usr/bin/apt-get -y install wget
RUN /usr/bin/apt-get -y install curl

RUN echo 'ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true' \
      | /usr/bin/debconf-set-selections
RUN echo 'msttcorefonts msttcorefonts/defoma select note' \
      | /usr/bin/debconf-set-selections

RUN /usr/bin/wget -O- \
      'https://chromium.googlesource.com/chromium/src/+/master/build/install-build-deps.sh?format=TEXT' \
        | /usr/bin/base64 -d > /tmp/install-build-deps.sh
RUN /usr/bin/wget -O- \
      'https://chromium.googlesource.com/chromium/src/+/master/build/install-build-deps-android.sh?format=TEXT' \
        | /usr/bin/base64 -d > /tmp/install-build-deps-android.sh
RUN /bin/chmod a+x /tmp/install-build-deps*.sh
# This runs install-build-deps.sh as well.
RUN /bin/bash -x /tmp/install-build-deps-android.sh

# Cleanup to shrink image sizes.
RUN /bin/rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/cache/man \
      /usr/share/{groff,info,lintian,linda,man}
RUN (/usr/bin/find /usr/share/doc -depth -type f ! -name copyright; \
     /usr/bin/find /usr/share/doc -empty) | /usr/bin/xargs /bin/rm -rf

RUN /bin/mkdir -p /b/swarm_slave

ADD start_swarm_bot.sh /opt/start_swarm_bot.sh
RUN /bin/chmod a+x /opt/start_swarm_bot.sh

CMD ["/opt/start_swarm_bot.sh"]
