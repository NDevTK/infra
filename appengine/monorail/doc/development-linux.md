# Getting started with Monorail development on Linux

These steps should only need to be taken once when setting up a new development
machine:

1. Install basic build dependencies:
    1.  `sudo apt-get install git build-essential automake`
    1. [Install Chrome depot_tools](https://commondatastorage.googleapis.com/chrome-infra-docs/flat/depot_tools/docs/html/depot_tools_tutorial.html#_setting_up) and ensure it is in your `$PATH`.

1. Check out the Monorail source code repository
    1. `mkdir -p ~/src/chrome/`
    1. `cd ~/src/chrome/`
    1. `fetch infra` (Googlers may alternatively `fetch infra_internal`)
    1. `cd infra/appengine/monorail`
    1. `gclient runhook`
    1. `gclient sync`

1. Install and configure the AppEngine SDK:
    1. It should be fetched for you by `gclient runhooks` above, otherwise, follow https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Python
1. Configure the AppEngine SDK: https://cloud.google.com/appengine/docs/standard/python3/setting-up-environment
    1. `gcloud auth login`
    1. `gcloud config set project monorail-prod`

1. Install and configure Docker:
    1. `sudo apt install docker-ce` (Googlers will need to follow instructions at go/docker)
    1. `sudo systemctl start docker`
    1. `sudo adduser $USER docker`
    1. Logout and back in to pick up the group change.
    1. `docker run hello-world` to verify installation.
    1. `gcloud auth configure-docker`

1. Configure the MySQL database:
    1. `sudo apt install libmariadb-dev`
    1. `docker-compose -f dev-services.yml up -d mysql
    1. `docker cp schema/. mysql:/schema`
    1. `docker exec -it mysql -- mysql --user=root -e 'CREATE DATABASE monorail;'
    1. `docker exec -it mysql -- mysql --user=root monorail < schema/framework.sql`
    1. `docker exec -it mysql -- mysql --user=root monorail < schema/project.sql`
    1. `docker exec -it mysql -- mysql --user=root monorail < schema/tracker.sql`
    1. `docker exec -it mysql -- mysql --user=root monorail -e "UPDATE User SET is_site_admin = TRUE WHERE email LIKE '%@example.com';"`

1. Install Monorail backend (python) prerequisites:
    1. `sudo apt install python2.7 python2.7-dev python-is-python2` (Googlers
       will need to run `sudo apt-mark hold python2 python-is-python2
       python2-dev` to prevent these from being uninstalled.)
    1. `curl -o /tmp/get-pip.py https://bootstrap.pypa.io/pip/2.7/get-pip.py`
    1. `sudo python2.7 /tmp/get-pip.py`
    1. `pip install virtualenv`
    1. `python2.7 -m virtualenv venv`

1. Install Monorail frontend (node) prerequisites:
    1. `curl https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash`
    1. `source "$HOME/.nvm/nvm.sh"`
    1. `nvm install 12.13.1`

These steps will need to be repeated regularly when developing Monorail:

1. Enter the virtual environment (in every new terminal you create):
    1. `eval $(../../go/env.py)`
    1. `source ./venv/bin/activate`

1. Install per-language dependencies (every few days to ensure you're up to date):
    1. `make dev_deps`
    1. `make js_deps`
    1. `npm install`

1. Launch the Monorail server (once per machine/per reboot):
    1. `make serve`
    1. Browse the app at localhost:8080 your browser.

1. Login with test user (once per machine or when starting with a fresh database):
       1.  Sign in using `test@example.com`
       1.  Log back out and log in again as `example@example.com`
       1.  Log out and finally log in again as `test@example.com`.
