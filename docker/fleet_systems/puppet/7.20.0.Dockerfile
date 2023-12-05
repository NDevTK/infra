FROM ubuntu:22.04

ENV DEBIAN_FRONTEND noninteractive
ENV PATH /opt/puppetlabs/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update && \
    apt-get dist-upgrade -y && \
    apt-get install -y curl && \
    curl -o /tmp/puppetlabs_release.deb https://apt.puppet.com/puppet7-release-jammy.deb && \
    dpkg -i /tmp/puppetlabs_release.deb && \
    rm -f /tmp/puppetlabs_release.deb && \
    apt-get update && \
    apt-get install -y puppet-agent=7.20.0-1jammy

RUN apt-get autoremove -y && \
    apt-get autoclean -y && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /var/cache/man \
      /usr/share/{doc,groff,info,lintian,linda,man}
