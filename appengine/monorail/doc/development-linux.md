# Getting started with Monorail development on Linux

These steps should only need to be taken once when setting up a new development
machine:

1. Install basic build dependencies:
    1.  `sudo apt-get install git build-essential automake`
    1. [Install Chrome depot_tools](https://commondatastorage.googleapis.com/chrome-infra-docs/flat/depot_tools/docs/html/depot_tools_tutorial.html#_setting_up) and ensure it is in your `$PATH`.

1. Check out the Monorail source code repository:
    1. `mkdir ~/src/`
    1. `cd ~/src/`
    1. `fetch infra` (Googlers may alternatively `fetch infra_internal`, but note that `gclient sync` will then always sync based on the current `infra_internal` branch instead of the current `infra` branch.)
    1. `cd infra/appengine/monorail`

1. Install and configure the gcloud CLI:
    1. It should be fetched for you by `fetch infra` above under `~/src/gcloud/bin`. Add it to your `PATH` in `.bashrc`.
    1. Otherwise, follow https://cloud.google.com/sdk/docs/install (Googlers on Linux may use `sudo apt-get install -y google-cloud-sdk`)
1. Configure gcloud:
    1. `gcloud auth login`
    1. `gcloud config set project monorail-dev`

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

1. Install Monorail frontend (node) prerequisites:
    1. `curl https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash`
    1. `source "$HOME/.nvm/nvm.sh"`
    1. `nvm install 12.13.1`

These steps will need to be repeated regularly when developing Monorail:

1. Enter the virtual environment (in every new terminal you create):
    1. `eval $(../../go/env.py)`
    1. `source ./venv/bin/activate`

1. Install per-language dependencies (every few days to ensure you're up to date):
    1. `make jsdeps`

1. Launch the Monorail server (once per machine/per reboot):
    1. `make serve`
    1. Browse the app at localhost:8080 your browser.

1. Login with test user (once per machine or when starting with a fresh database):
       1.  Sign in using `test@example.com`
       1.  Log back out and log in again as `example@example.com`
       1.  Log out and finally log in again as `test@example.com`.
