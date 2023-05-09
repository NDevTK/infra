# Monorail Development Problems

*   `BindError: Unable to bind localhost:8080`

This error occurs when port 8080 is already being used by an existing process. Oftentimes,
this is a leftover Monorail devserver process from a past run. To quit whatever process is
on port 8080, you can run `kill $(lsof -ti:8080)`.

*   `gcloud: command not found`

Add the following to your `~/.zshrc` file: `alias gcloud='/Users/username/google-cloud-sdk/bin/gcloud'`. Replace `username` with your Google username.

*   `TypeError: connect() argument 6 must be string, not None`

This occurs when your mysql server is not running.  Check if it is running with `ps aux | grep mysqld`.  Start it up with <code>/etc/init.d/mysqld start </code>on linux, or just <code>mysqld</code>.

*   dev_appserver says `OSError: [Errno 24] Too many open files` and then lists out all source files`

dev_appserver wants to reload source files that you have changed in the editor, however that feature does not seem to work well with multiple GAE modules and instances running in different processes.  The workaround is to control-C or `kill` the dev_appserver processes and restart them.

*   `IntegrityError: (1364, "Field 'comment_id' doesn't have a default value")` happens when trying to file or update an issue

In some versions of SQL, the `STRICT_TRANS_TABLES` option is set by default. You'll have to disable this option to stop this error.

# Development resources

## Supported browsers

Monorail supports all browsers defined in the [Chrome Ops guidelines](https://chromium.googlesource.com/infra/infra/+/main/doc/front_end.md).

File a browser compatibility bug
[here](https://bugs.chromium.org/p/monorail/issues/entry?labels=Type-Defect,Priority-Medium,BrowserCompat).

## Frontend code practices

See: [Monorail Frontend Code Practices](doc/code-practices/frontend.md)

## Monorail's design

* [Monorail Data Storage](doc/design/data-storage.md)
* [Monorail Email Design](doc/design/emails.md)
* [How Search Works in Monorail](doc/design/how-search-works.md)
* [Monorail Source Code Organization](doc/design/source-code-organization.md)
* [Monorail Testing Strategy](doc/design/testing-strategy.md)

## Triage process

See: [Monorail Triage Guide](doc/triage.md).

## Release process

See: [Monorail Deployment](doc/deployment.md)

# User guide

For information on how to use Monorail, see the [Monorail User Guide](doc/userguide/README.md).

## Setting up a new instance of Monorail

See: [Creating a new Monorail instance](doc/instance.md)
