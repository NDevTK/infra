# ChOpsUI Demo Application

This is a minimal app that demonstrates:
- typescript
- webpack dev server
- sourcemaps
- protobuf support
- redux in TypeScript
- @chopsui npm package components

TODO:
- Unit tests
- webpack for deployment

## Quick Start
```
npm ci
npm run build:proto
npm run start:dev
```

Then point your browser at the `Project is running at` URL it spits out to the console, e.g. http://localhost:8082

The `start:dev` command will run a webpack devserver in "watch" mode, so changes you
make to any of the source files should automatically be reprocessed (compiled, bundled etc)
and the browser window you have open to the devserver should automatically reload itself.

## TypeScript

`ts` is pretty noisy and generates a lot of other files alongside your source, so we ignore it for git purposes (see `.gitignore` for details).

- `*.d.ts` are type declaration files. These are kind of like header files. Only `tsc` should consume them.
- `*.ts.map` and `*.js.map` are sourcemap files. Indespensable for interactive debugging, but otherwise ignorable. See Source Maps section below.
- `.tsbuildinfo` some diagnostic info about your compile.

You can run `npm run lint` to run the linter over the typescript code in this project.
This would be a good presubmit check, actually.

TODO: Link to more formal TypeScript guidelines for ChOps Apps.

### LitElement properties

TS has type annotation features that will let you declare LitElement properties
more compactly:

```
export class Demo extends connect<AppState>(store) LitElement {
  @property() _privateProp: string = 'foo';
  ...
  stateChanged(state: AppState) {
    this._todos = state.todos || [];
    // The above statement automatically schedules a render() call
    // to update the html.
  }
}
```

## WebPack

The `npm run start:dev` command is great for debugging and local testing, but it doesn't produce anything you can deploy to prod or other environments.

For that, there's a `npm run build` command, and it generates minified and bundled js in the ./dist directory.

TODO: Say much more on this configuration.  Many parts of it are subtle and break easily.

## Source Maps

It's a miracle that this works at all, given the house of cards upon which it rests, but you can indeed open up Chrome Dev Tools and set breakpoints
in the original TypeScript code.  Look for the "top/webpack://" tree in the "Sources" tab of devtools. Depending on your source layout, the original
.ts file may be under different sub-trees so keep digging if you don't see it right away.

## Protocol Buffers

The `npm run build:proto` command will run protoc to generate JS and TypeScipt bindings for
monorail's API definitions (referenced as relative paths).

This will place the generated typescript files in `./generated` and create the directory first if
necessary. This directory is git ignored because it can be entirely reconstructed from
existing source files.

Note that you have to have `protoc` installed before this will work. That's beyond this README
file, but if you're working on ChOps code you should already have it in your environment.

TODO: Get webpack to automatically run protoc when .proto files change.

Note that [grpc-web]() doesn't quite work with prpc servers right out of the box.  The demo
application offers one tweak so far: rename grpcweb-generated FooList attributes to just Foo.

E.g. `ListIssuesRequest` has a `repeated string project_names` field
which grpc-web stubs out with a setter named `setProjectNamesList` and if you call `toObject` on an
instance of this class, you'll indeed get a property named `projectNamesList` which confuses the monorail
pRPC API, which expects it to just be called `projectNames`.

Hence this fixer function in demo.ts:

```
 _fixArrayNames(obj) {
    const names = Object.getOwnPropertyNames(obj)
    names.forEach((name) => {
      if (name.endsWith('List')) {
        const p = obj[name]
        delete obj[name]
        obj[name.substr(0, name.length-4)] = p
      }
    })
    return obj
 }
```
Otherwise, I believe we can use these message types.  Just ignore the grpc service bindings,
and stick to calling your pPRC servers the ChOps way:

```

import {ListStatusesRequest} from './generated/api/api_proto/projects_pb'

...

const req = new ListStatusesRequest()
req.setProjectName('test') // This gets type checked by tsc.

// This assumes you're running in dev mode w/ MR server on localhost.
const prpcClient = new PrpcClient({host:monorailHost, insecure:true})
// Convert the grpc-web proto type into a plain object we can send via prpc as json.
const fixedReq = this._fixArrayNames(req.toObject())

try {
  // Call prcpClient.call like you normall would:
  const resp = await prpcClient.call('monorail.Projects', 'ListStatuses', fixedReq)
  console.log(resp.getStatusDefsList()) // This is type checked too.
} catch (e) {
  ...
}
```

Note that since prcpClient isn't yet type-aware, we can't yet use RPC response types.
TODO: Fix that.

## Redux with TypeScript

Types can help out with redux patterns. [Redux.js.org's TypeScript advice is here](https://redux.js.org/recipes/usage-with-typescript).

Application state and actions are obvious places to start adding types.
This can help catch typos and potentially undefined values at compile time.

### Application State

Your application state can be just a plain interface type.  It doesn't
need a base class or type to extend.

```
export interface AppState {
  title: string
  user: User
  ...
}
```

### Actions

Redux itself provides some helper types, like `Action`, which
we can extend to cusomize for our use cases. Avoid using the
`AnyAction` type as we discourage the use of the `any` type elsewhere
in your application.

```
import {Action} from 'redux';

interface UpdateUserAction extends Action {
  type: 'UPDATE_USER'
  user: User
}
```

### Putting Redux types together

```
function rootReducer(state: AppState = {}, action:AppAction): AppState {
  return {
    title: title(state.title)
  }
}
```

## Protocol buffers and Redux

Don't.

It may be tempting to use protoc-generated types to structure your
application state, but be warned: they don't de/serialze properly in this
context. I know, the irony of protos not serializing...

