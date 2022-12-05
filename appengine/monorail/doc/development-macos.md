Here's how to run Monorail locally for development on MacOS and Debian stretch/buster or its derivatives.

1.  You need to [get the Chrome Infra depot_tools commands](https://commondatastorage.googleapis.com/chrome-infra-docs/flat/depot_tools/docs/html/depot_tools_tutorial.html#_setting_up) to check out the source code and all its related dependencies and to be able to send changes for review.
1.  Check out the Monorail source code
    1.  `cd /path/to/empty/workdir`
    1.  `fetch infra` (make sure you are not "fetch internal_infra" )
    1.  `cd infra/appengine/monorail`
1.  Make sure you have the AppEngine SDK:
    1.  It should be fetched for you by step 1 above (during runhooks)
    1.  Otherwise, you can download it from https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Python
    1.  Also follow https://cloud.google.com/appengine/docs/standard/python3/setting-up-environment to setup `gcloud`
1.  Install CIPD dependencies:
    1. `gclient runhooks`
1.  Install MySQL v5.6.
    1. On Mac, use [homebrew](https://brew.sh/) to install MySQL v5.6:
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
        1.  ``eval `../../go/env.py` `` -- you'll need to run this in any shell you
            wish to use for developing Monorail. It will add some key directories to
            your `$PATH`.
        1.  Install build requirements:
            1.  `sudo apt-get install build-essential automake`
    1. On MacOS
        1. [Install homebrew](https://brew.sh)
        1.  Install node and npm
            1.  Install node version manager `brew install nvm`
            1.  See the brew instructions on updating your shell's configuration
            1.  Install node and npm `nvm install 12.13.0`
            1.  Add the following to the end of your `~/.zshrc` file:

                    export NVM_DIR="$HOME/.nvm"
                    [ -s "/usr/local/opt/nvm/nvm.sh" ] && . "/usr/local/opt/nvm/nvm.sh"  # This loads nvm
                    [ -s "/usr/local/opt/nvm/etc/bash_completion.d/nvm" ] && . "/usr/local/opt/nvm/etc/bash_completion.d/nvm"  # This loads nvm bash_completion

1.  Install Python and JS dependencies:
    1.  Optional: You may need to install `pip`. You can verify whether you have it installed with `which pip`.
        1. make sure to install `pip` using `python2` instead of `python3` (use `python --version` to check the version for 2.7, `which python2` to check the path)
            1. If you need python 2.7 for now: `sudo apt install python2.7 python2.7-dev python-is-python2`
        1. `curl -O /tmp/get-pip.py https://bootstrap.pypa.io/pip/2.7/get-pip.py`
        1. `sudo python /tmp/get-pip.py`
    1.  Use `virtualenv` to keep from modifying system dependencies.
        1. `pip install virtualenv`
        1. `python -m virtualenv venv` to set up virtualenv within your monorail directory.
        1. `source venv/bin/activate` to activate it, needed in each terminal instance of the directory.
    1.  Mac only: install [libssl](https://github.com/PyMySQL/mysqlclient-python/issues/74), needed for mysqlclient. (do this in local env not virtual env)
        1. `brew install openssl`
    1.  `LIBRARY_PATH=/usr/local/opt/openssl/lib LDFLAGS="-L/usr/local/opt/openssl/lib" CPPFLAGS="-I/usr/local/opt/openssl/include" make dev_deps` (run in virtual env)
    1.  `make jsdeps`
1.  Run the app:
    1.  `make serve` (run in virtual env)
    1.  Start MySQL:
        1. Mac: `brew services restart mysql@5.6`
        1. Linux: `mysqld`
1. Browse the app at localhost:8080 your browser.
1. Set up your test user account (these steps are a little odd, but just roll with it):
       1.  Sign in using `test@example.com`
       1.  Log back out and log in again as `example@example.com`
       1.  Log out and finally log in again as `test@example.com`.
       1.  Everything should work fine now.
1.  Modify your Monorail User row in the database and make that user a site admin. You will need to be a site admin to gain access to create projects through the UI.
    1.  `mysql --user=root monorail -e "UPDATE User SET is_site_admin = TRUE WHERE email = 'test@example.com';"`
    1.  If the admin change isn't immediately apparent, you may need to restart your local dev appserver. If you kill the dev server before running the docker command, the restart may not be necessary.