The script here is used to manage the Docker containers on swarming bots. By
design, it's stateless and event-driven. Currently it has only one mode of
execution: "launch". This mode ensures that there is always a given number of
swarming containers running on the bot. On the bots, the script called in this
mode every 5 minutes via cron.

It's intended to be run in conjucture with the swarm_docker container image.
More information on the image can be found
[here](https://chromium.googlesource.com/infra/infra/+/HEAD/docker/swarm_docker/README.md).


Launching the containers
--------------------------
Every 5 minutes, this script is invoked with the `launch` argument. This is
how the containers get spawned, shut down, and cleaned up. Essentially, it does
the following:
* Gracefully shutdowns each container if container uptime is too high (default
4 hours) or host uptime is too high (default 24 hours).
* Verifies that required number of containers is running and spawns new ones as
  needed.


Gracefully shutting down a container
--------------------------
To preform various forms of maintenance, the script here gracefully shuts down
containers and, by association, the swarming bot running inside them. This is
done by sending SIGTERM to the swarming bot process. This alerts the swarming
bot that it should quit at the next available opportunity (ie: not during a
test.)
In order to fetch the pid of the swarming bot, the script runs `lsof` on the
[swarming.lck](https://cs.chromium.org/chromium/infra/luci/appengine/swarming/doc/Bot.md?rcl=8b90cdd97f8f088bcba2fa376ce49d9863b48902&l=305)
file.


Fetching the docker image
--------------------------
When a new container is launched with an image that is missing
on a bot's local cache, it pulls it from the specified container registry. By
default, this is the gcloud registry
[chops-public-images-prod](https://console.cloud.google.com/gcr/images/chops-public-images-prod/global/swarm_docker).
These images are world-readable, so no authentication with the registry is
required. Additionally, when new images are fetched, any old and unused images
still present on the machine will be deleted.


File locking
--------------------------
To avoid multiple simultaneous invocations of this service from stepping on
itself, parts of the script interacting with containers are wrapped in a mutex
via a flock. Each container has its own lock file.

Deploying the script
--------------------------
The script and its dependencies are deployed as a CIPD package via puppet. The
package (infra/swarm_docker_tools/$platform) is continously built on this
[bot](https://ci.chromium.org/p/infra-internal/builders/luci.infra-internal.prod/infra-packager-linux-64).
Puppet deploys it to the relevant bots at
[these revisions](https://chrome-internal.googlesource.com/infra/puppet/+/94c1a63589e03a543035937b586dbf86ba5b2d3e/puppetm/etc/puppet/hieradata/cipd.yaml#417).
The canary pin, which tracks latest version in the repository, currently
affects no bots on
[chromium-swarm-dev](https://chromium-swarm-dev.appspot.com).
So you can proceed to update the stable pin immediately after making a change.

The call site of the script is also defined in
[puppet](https://chrome-internal.googlesource.com/infra/puppet/+/94c1a63589e03a543035937b586dbf86ba5b2d3e/puppetm/etc/puppet/modules/chrome_infra/templates/setup/docker/swarm/swarm_docker_cron.sh.erb).
