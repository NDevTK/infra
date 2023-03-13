FROM debian:bookworm-slim

RUN apt-get update

#Install virtme
RUN apt-get -y install python3 git python3-setuptools
RUN cd /tmp; git clone git://git.kernel.org/pub/scm/utils/kernel/virtme/virtme.git
RUN cd /tmp/virtme ; ./setup.py install

#Virtme running deps
RUN apt-get install -y qemu-system-x86 busybox-static udev kmod
RUN mkdir -p /lib/modules

#Virtme configkernel deps
RUN apt-get install -y make gcc flex bison

#kunit deps
RUN apt-get install -y bc libelf-dev

#Kernel build
RUN apt-get install -y make gcc gcc-aarch64-linux-gnu flex bison sparse libssl-dev bc git libelf-dev autoconf libtool
