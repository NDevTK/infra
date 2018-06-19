# tricium/appengine/frontend/ui

This directory contains code for the web UI for the Tricium service.

## Development

## Setup

Run `npm install` to install all dependencies.

### Testing

Run `make test` to run the tests from the command line.

### Building

Run `polymer build` before deployment, or before serving via the
App Engine dev server.

### Linting

Run `polymer lint`; additionally run eslint on any changed files.

### Local development

Run `polymer serve` and navigate to http://localhost:8081.  Local
changes should be immediately available after refreshing the page.
