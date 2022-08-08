#!/bin/bash
set -e

# Note: this is partially based on infra/docker/docker_devices/Dockerfile

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get -y dist-upgrade
apt-get -y install python3 python3-distutils sudo tzdata ca-certificates

# The user that will run botholder and the swarming bot itself.
useradd --create-home --shell /bin/bash --user-group swarming

# Where Swarming bot code will live.
mkdir /b
chown swarming:swarming /b

# Where oauth2_access_token_config.json will live.
mkdir -p /etc/swarming_config
chown swarming:swarming /etc/swarming_config

# Replace /sbin/shutdown used by Swarming bot to reboot the host. Instead it
# will just signal `botholder` to restart the Swarming process.
rm -f /sbin/shutdown
cp ./shutdown.sh /sbin/shutdown
chmod 0755 /sbin/shutdown

# Allow swarming to call it via passwordless sudo, since it does so.
echo "%swarming ALL = NOPASSWD: /sbin/shutdown" > /etc/sudoers.d/swarming

# See os_utilities.get_hostname(). Without this, the bot uses GKE **node**
# hostname as Swarming bot ID, not the pod name.
touch /.dockerenv
