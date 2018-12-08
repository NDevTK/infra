# chopsui

This directory contains Web Components that are meant to be shared
by Chrome Operations' application frontends.

ChopsUI packages are published publicly on NPM under the  [@chopsui](https://www.npmjs.com/settings/chopsui/packages)
organization. To be added to this org, please send your NPM username to zhangtiff@.

# Publishing packages

TODO(zhangtiff): Streamline the process of publishing multiple packages, especially once we add more packages.

Packages are currently independently published from each other. Thus, in order to publish/update an individual package, move into the directory for the package you want to update and publish the package.

For example, in order to update chops-header, you would run:

```sh
cd elements/chops-header
npm publish --access public
```

You will have to login to an NPM account with access to the @chopsui org to have publish permissions.

# Testing demos locally

TODO(zhangtiff): Replace polymer-cli with Webpack.

To run demos locally, we use polymer-cli to serve static files. Polymer-cli does path rewriting under the hood to allow us to use named imports for ES modules without building but eventually we want to switch to build components with Webpack.

To run a demo, run:

```sh
polymer serve
```

Them navigate to your demo URL. For example, for chops-header this URL is:

```sh
http://localhost:8001/demos/chops-header.html
```

# Downloading packages

Install the name of the specific package you're looking and install it within the @chopsui package scope. For example, to install chops-header, run:

```sh
npm install @chopsui/chops-header
```
