# Monorail Issue Tracker

Monorail is the Issue Tracker used by the Chromium project and other related
projects. It is hosted at [bugs.chromium.org](https://bugs.chromium.org).

If you wish to file a bug against Monorail itself, please do so in our
[self-hosting tracker](https://bugs.chromium.org/p/monorail/issues/entry).
We also discuss development of Monorail at `infra-dev@chromium.org`.

# Getting started with Monorail development

*For Googlers:* Monorail's codebase is open source and can be installed locally on your workstation of choice.

For local development on Linux, see [Linux development instructions](doc/development-linux.md)

Instructions for deploying Monorail to an existing instance or setting up a new instance are [here](doc/deployment.md).

See also: [Common Development Problems](doc/development-problems.md)

## Feature Launch Tracking

To set up FLT/Approvals in Monorail:
1. Visit the gear > Development Process > Labels and fields
1. Add at least one custom field with type "Approval" (this will be your approval
1. Visit gear > Development Process > Templates
1. Check "Include Gates and Approval Tasks in issue"
1. Fill out the chart - The top row is the gates/phases on your FLT issue and you can select radio buttons for which gate each approval goes

## Testing

To run all Python unit tests, in the `appengine/monorail` directory run:

```
make test
```

For quick debugging, if you need to run just one test you can do the following. For instance for the test
`IssueServiceTest.testUpdateIssues_Normal` in `services/test/issue_svc_test.py`:

```
../../test.py test appengine/monorail:services.test.issue_svc_test.IssueServiceTest.testUpdateIssues_Normal --no-coverage
```

### Frontend testing

To run the frontend tests for Monorail, you first need to set up your Go environment. From the Monorail directory, run:

```
eval `../../go/env.py`
```

Then, to run the frontend tests, run:

```
make karma
```

If you want to skip the coverage for karma, run:
```
make karma_debug
```

To run only one test or a subset of tests, you can add `.only` to the test
function you want to isolate:

```javascript
// Run one test.
it.only(() => {
  ...
});

// Run a subset of tests.
describe.only(() => {
  ...
});
```

Just remember to remove them before you upload your CL.

# Development resources

## Supported browsers

Monorail supports all browsers defined in the [Chrome Ops guidelines](https://chromium.googlesource.com/infra/infra/+/main/doc/front_end.md).

File a browser compatability bug
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

See: [Monorail Deployment](http://go/monorail-deploy)

# User guide

For information on how to use Monorail, see the [Monorail User Guide](doc/userguide/README.md).

## Setting up a new instance of Monorail

See: [Creating a new Monorail instance](doc/instance.md)
