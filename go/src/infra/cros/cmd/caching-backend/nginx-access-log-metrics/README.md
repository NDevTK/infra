## Tsmon setup

The implementation depends on the default tsmon configuration. The host
running/testing this code must have `/etc/chrome-infra/ts-mon.json` file with
the content as shown below:

```
{
  "credentials":"/path/to/service_account.json",
  "endpoint":"server URL",
  "use_new_proto": true
}
```
