#!/bin/bash

SWARM_DIR=/b/swarm_slave
SWARM_URL="https://chromium-swarm.appspot.com/bot_code"
SWARM_ZIP=swarming_bot.zip

/bin/chown chrome-bot:chrome-bot $SWARM_DIR
/bin/su -c "/usr/bin/curl -sSLOJ $SWARM_URL -o ${SWARM_DIR}/${SWARM_ZIP}" chrome-bot

echo "Starting $SWARM_ZIP"
cd $SWARM_DIR \
  && exec su -c \
    "/usr/bin/python $SWARM_ZIP start_bot" chrome-bot
