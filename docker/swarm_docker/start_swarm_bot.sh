#!/bin/bash -x

trap "exit 10" SIGUSR1

SWARM_DIR=/b/swarming
SWARM_ZIP=swarming_bot.zip
LUCI_MACHINE_TOKEN_FILE='/var/lib/luci_machine_tokend/token.json'

curl_header_args=()
if [[ -r "${LUCI_MACHINE_TOKEN_FILE}" ]]; then
  luci_token="$(jq -r ".luci_machine_token" "${LUCI_MACHINE_TOKEN_FILE}")"
  curl_header_args=("--header" "X-Luci-Machine-Token: ${luci_token}")
fi

mkdir -p $SWARM_DIR
/bin/chown chrome-bot:chrome-bot $SWARM_DIR
cd $SWARM_DIR
rm -rf swarming_bot*.zip
sudo -u chrome-bot -- /usr/bin/curl -sSLOJ "${curl_header_args[@]}" "${SWARM_URL}"

# Initialize tuntap for qemu running inside of the docker instance on arm64.
if [[ $(arch) = "aarch64" ]]; then
  ip tuntap add qemu mode tap user chrome-bot
  ip link set dev qemu up
  ip addr flush qemu
fi

echo "Starting $SWARM_ZIP"
# Run the swarming bot in the background, and immediately wait for it. This
# allows the signal trapping to actually work.
/bin/su -c "/usr/bin/python3 $SWARM_ZIP start_bot" chrome-bot &
wait %1
exit $?
