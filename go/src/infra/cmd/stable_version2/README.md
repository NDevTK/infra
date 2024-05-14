# stable_version2

For more info, see: go/stable_version

`stable_version2` is a script to validate ChromeOS lab stable_version [config](https://chrome-internal.googlesource.com/chromeos/infra/config/+/main/lab_platform/generated/stable_versions.cfg).

## Automated testing

To run the tests for `stable_version2`, run:

`make test`

## Manual testing

First, build `stable_version2`:

`make stable_version2`

Then, run: `./stable_version2 validate-config
~/chromiumos/infra/config/lab_platform/generated/stable_versions.cfg`

`stable_versions.cfg` is available as
[part of the ChromiumOS repo](https://chrome-internal.googlesource.com/chromeos/infra/config/+/main/lab_platform/generated/stable_versions.cfg).
For the above command, the exact path to your local copy of
`stable_versions.cfg` might be different depending on how your folders are set
up. You can also copy the `stable_versions.cfg` file somewhere else and manually
make changes for testing.

If you need to login, run:

`./stable_version2 login`
