Here's how to run Monorail locally for development on MacOS and Debian stretch/buster or its derivatives.

1.  You need to [get the Chrome Infra depot_tools commands](https://commondatastorage.googleapis.com/chrome-infra-docs/flat/depot_tools/docs/html/depot_tools_tutorial.html#_setting_up) to check out the source code and all its related dependencies and to be able to send changes for review.
1.  Check out the Monorail source code
    1.  `cd /path/to/empty/workdir`
    1.  `fetch infra`
    1.  `cd infra/appengine/monorail`
1.  Set up the gcloud CLI
    1.  It should be fetched for you by the previous step (during `gclient runhooks`)
    1.  Add `workdir/gcloud/bin` to `PATH` in `.zshrc`
    1. `gcloud auth login`
    1. `gcloud auth application-default login`
    1. `gcloud config set project monorail-dev`
1.  Install MySQL v5.6.
    1. Install [homebrew](https://brew.sh/)
    1.  `brew install mysql@5.6`
    1. Otherwise, download from the [official page](http://dev.mysql.com/downloads/mysql/5.6.html#downloads).
        1.  **Do not download v5.7 (as of April 2016)**
1.  Set up SQL database. (You can keep the same sharding options in settings.py that you have configured for production.).
    1. Copy setup schema into your local MySQL service.
        1.  `mysql --user=root -e 'CREATE DATABASE monorail;'`
        1.  `mysql --user=root monorail < schema/framework.sql`
        1.  `mysql --user=root monorail < schema/project.sql`
        1.  `mysql --user=root monorail < schema/tracker.sql`
        1.  `exit`
1.  Configure the site defaults in settings.py.  You can leave it as-is for now.
1.  Set up the front-end development environment:
    1. On Debian
        1.  Install build requirements:
            1.  `sudo apt-get install build-essential automake`
    1. On MacOS
        1. [Install homebrew](https://brew.sh)
        1.  Install npm
            1.  Install node version manager `brew install nvm`
            1.  See the brew instructions on updating your shell's configuration
            1.  Install node and npm `nvm install 12.13.0`
            1.  Add the following to the end of your `~/.zshrc` file:

                    export NVM_DIR="$HOME/.nvm"
                    [ -s "/usr/local/opt/nvm/nvm.sh" ] && . "/usr/local/opt/nvm/nvm.sh"  # This loads nvm
                    [ -s "/usr/local/opt/nvm/etc/bash_completion.d/nvm" ] && . "/usr/local/opt/nvm/etc/bash_completion.d/nvm"  # This loads nvm bash_completion

1.  Install JS dependencies:
    1.  `make jsdeps`
1.  Run the app:
    1.  Start MySQL:
        1. Mac: `brew services restart mysql@5.6`
        1. Linux: `mysqld`
    1.  `make serve`
1. Browse the app at localhost:8080 your browser.
1. Create your test user accounts
       1.  Sign in using `test@example.com`
       1.  Log back out and log in again as `example@example.com`
       1.  Log out and finally log in again as `test@example.com`
       1.  Everything should work fine now.
1.  Modify your Monorail User row in the database and make that user a site admin. You will need to be a site admin to gain access to create projects through the UI.
    1.  `mysql --user=root monorail -e "UPDATE User SET is_site_admin = TRUE WHERE email = 'test@example.com';"`
    1.  If the admin change isn't immediately apparent, you may need to restart your local dev appserver. If you kill the dev server before running the docker command, the restart may not be necessary.
