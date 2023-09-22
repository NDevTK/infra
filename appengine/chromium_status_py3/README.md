# Chromium Status

Base framework for:

* https://chromium-status.appspot.com/
* https://chromiumos-status.appspot.com/
* See [tools/update.py] for the complete list.

## Python 3

This is an effort to migrate chromium status to Python3 running on GAE v2. The
old source code (Python 2) is at https://source.corp.google.com/chromium_infra/appengine/chromium_status/.
The bug tracking the effort is https://crbug.com/1121016.

The effort was put on hold due to deprioritisation. So far the following
components have been migrated:

- Updated the runtime environment to py3
- Updated app.yaml to be py3-compatibile
- Created main.py as the entry point
- Enabled Cloud Logging
- Migrated google.appengine.api.db to Cloud NDB
- Migrate users service to OAuth2 + OpenID Connect

There are still items that need to be done:
- Authorization for service accounts (e.g. luci-notify, gerrit)
- Memcache -> Cloud Memory store (if it turns out that memcache is necessary)
- Test and deploy

## Developing

You should download & install Google Appengine into the `appengine_module/` dir:

https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Python

Unzip it there so that `google_appengine/` exists alongside `chromium_status/`.

You can launch a local instance for testing by doing:

```
$ ./appengine_module/google_appengine/dev_appserver.py --host $(hostname -s) app.yaml
```

Then browse to your desktop's port 8080.

To post commits with git, you can do:

```
... make changes ...
$ git commit -a -m 'commit message'
$ git cl upload
```

Integration tests found in tests/ use dev_appserver.py
but aren't allowed to actually load/test python code directly.

We should generalize the mechanism for writing dev_appserver.py tests
and share that with other appengine_apps.
