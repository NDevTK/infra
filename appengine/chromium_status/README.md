# Chromium Status

Base framework for:

* https://chromium-status.appspot.com/
* https://chromiumos-status.appspot.com/
* See [tools/update.py](./tools/update.py) for the complete list.

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
