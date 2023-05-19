The files here are used to build a docker image suitable for sandboxing a USB
or network device into its own execution environment, in which a swarming bot is
running. More information is at
[go/one-device-swarming](http://go/one-device-swarming)


The docker image
--------------------------
The `build.sh` script will create two local images named `android_docker`
and `cros_docker` tagged with the date and time of creation. The image itself is
simply an Ubuntu flavor (xenial as of writing this) with a number of packages
and utilities installed, mainly via chromium's
[install_build_deps.py](https://chromium.googlesource.com/chromium/src/+/HEAD/build/install-build-deps.py).
When launched as a container, this image is configured to run the
[start_swarm_bot.sh](https://chromium.googlesource.com/infra/infra/+/HEAD/docker/docker_devices/android/start_swarm_bot.sh)
script here which fetches and runs the bot code of the swarming server pointed
to by the `$SWARM_URL` env var. Note that running the image locally on a
developer workstation is unsupported.

### Automatic image building
Everyday at 8am PST, builders on the internal.infra.cron waterfall build a
fresh version of the images. These builders essentially run `./build.sh` and
upload the resultant images to a docker
[container registry](https://docs.docker.com/registry/). The registry, hosted
by gcloud, is located at
[chops-public-images-prod](https://console.cloud.google.com/gcr/images/chops-public-images-prod/global/).
Specifically, the [android builder](https://ci.chromium.org/p/infra-internal/builders/prod/android-docker-image-builder)
uploads its images [here](https://console.cloud.google.com/gcr/images/chops-public-images-prod/global/android_docker),
and the [ChromeOS builder](https://ci.chromium.org/p/infra-internal/builders/prod/cros-docker-image-builder)
uploads its images [here](https://console.cloud.google.com/gcr/images/chops-public-images-prod/global/cros_docker).

### Shutting container down from within
Because a swarming bot may trigger a reboot of the bot at any time (see
[docs](https://cs.chromium.org/chromium/infra/luci/appengine/swarming/doc/Magic-Values.md?rcl=8b90cdd97f8f088bcba2fa376ce49d9863b48902&l=65)),
it should also be able to shut its container down from within. By conventional
means, that's impossible; a container has no access to the docker engine
running outside of it. Nor can it run many of the utilities in `/sbin/` since
the hardware layer of the machine is not available to the container. However,
because a swarming bot uses the `/sbin/shutdown` executable to reboot, this
file is replaced with our own bash script here
([shutdown.sh](https://chromium.googlesource.com/infra/infra/+/HEAD/docker/docker_devices/android/shutdown.sh))
that simply sends SIGUSR1 to `init` (pid 1). For a container, `init` is whatever
command the container was configured to run (as opposed to actual `/sbin/init`).
In our case, this is [start_swarm_bot.sh](https://chromium.googlesource.com/infra/infra/+/HEAD/docker/docker_devices/android/start_swarm_bot.sh),
which conveniently traps SIGUSR1 at the very beginning and exits upon catching
it. Consequently, running `/sbin/shutdown` from within a container will cause
the container to immediately shutdown.


Image deployment
--------------------------
Image deployment is done via puppet. When a new image needs to be rolled out,
grab the image name in the step-text of the "Push image" step on the
[image-building bot](https://ci.chromium.org/p/infra-internal/builders/prod/android-docker-image-builder).
It should look like android_docker:$date. (The bot runs once a day, but you can
manually trigger a build if you don't want to wait.) To deploy, update the
image pins in puppet for [canary bots](https://chrome-internal.googlesource.com/infra/puppet/+/94c1a63589e03a543035937b586dbf86ba5b2d3e/puppetm/opt/puppet/conf/nodes.yaml#444),
followed by [stable bots](https://chrome-internal.googlesource.com/infra/puppet/+/94c1a63589e03a543035937b586dbf86ba5b2d3e/puppetm/opt/puppet/conf/nodes.yaml#587).
The canary pin affects bots on [chromium-swarm-dev](https://chromium-swarm-dev.appspot.com),
which the android testers on the [chromium.dev](https://luci-milo-dev.appspot.com/p/chromium/g/chromium.dev/console)
waterfall run tests against. If the canary has been updated, the bots look fine,
and the tests haven't regressed, you can proceed to update the stable pin.
(Note that it may take several hours for the image pin update to propagate. It's
advised to wait at least a day to update stable after updating canary.)

Launching the containers
------------------------
On the bots, container launching and managing is controlled via the python
service [here](https://chromium.googlesource.com/infra/infra/+/HEAD/infra/services/android_docker/).
The service has two modes of invocation:
* launch: Ensures every locally connected android device has a running
          container. Will gracefully tear down containers that are too old or
          reboot the host. Called every 5 minutes via cron.
* add_device: Give's a device's container access to its device. Called every
              time a device appears/reappears on the system via udev.

More information can be found [here](https://chromium.googlesource.com/infra/infra/+/HEAD/infra/services/android_docker/README.md).
