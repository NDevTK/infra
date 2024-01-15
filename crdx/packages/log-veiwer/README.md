# Log-viewer component

A component library that provides UI to display and handle logs.
It was built using [ReactJS](https://react.dev/),
[MUI](https://mui.com/), and [Virtusoo](https://virtuoso.dev/).

## Installation

1. Install the required dependencies:

   ```sh
   npm install react react-dom @mui/material @emotion/react @emotion/styled
   ```

2. Install via npm:

   ```sh
   npm install @chopsui/log-viewer
   ```

## Local development

There are two methods to locally test the component:

1. Testing the component with [Storebook](https://storybook.js.org/).

2. Testing the component's integration with your project.

### Using Storybook

* You can build log-viewer's components and see how look and act
  in isolation using [Storybook](https://storybook.js.org/).

* An example of this works is included under `src/base_component/base_component.stories.tsx`.

### Integrating the component with your project

In order to integrate the component with your project we will use
[npm-link](https://docs.npmjs.com/cli/v9/commands/npm-link).
NPM-link creates a symbolic link for the project in the global NPM repository.

**Set up local development:**

1. At the top directory of `log-viewer` run:

   ```sh
   npm link
   ```

   This will create a link to your local npm library copy in your global node_modules.

2. Then at your UI project's top directory run:

   ```sh
   npm link @chopsui/log-viewer
   ```

   This will use the linked version of the log-viewer from your global npm_modules.

**Clean up after local development:**

1. Only run this after you have finished your work and _after_ publishing
   the changes to npm registry.

2. Update the version of `@chopsui/log-viewer` in your package.json
   to the one you have published.

3. At your project's top directory run:

   ```sh
   npm ci @chopsui/log-viewer
   ```

   This will install the version you just published and undo the linking.

This will not remove the link we created in the global `node_modules`,
which allows us to reuse this link later, rather this will only remove
it from the project.

If you want to delete the global link as well, run this in the top
directory of `log-viewer`:

```sh
npm unlink -g
```

If you are not yet ready to use the published version, you can skip step
2 and just run step 3, which reset the version to the one you had before linking.
