# tricium/appengine/frontend/ui

This directory contains code for the web UI for the Tricium service.

## Development

## Setup

Run `npm install` to install all dependencies.

### Testing

Run `wct` or `polymer test` (they should both behave similarly).

### Building

Run `polymer build` before deployment, or before serving via the
App Engine dev server.

### Linting

Run `polymer lint`; additionally run `eslint` on any changed files.

### Local development

To run a local server with just the UI, run `polymer serve`.
This allows you to incrementally test changes without rebuilding.

To run a devserver with all endpoints (not just UI, run
`polymer build` and then `gae.py devserver`.
