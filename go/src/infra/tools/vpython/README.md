## vpython - simple and easy VirtualEnv Python

`vpython` is a tool, written in Go, which enables the simple and easy invocation
of Python code in [VirtualEnv](https://virtualenv.pypa.io/en/stable/)
environments.

For the standard case, employing `vpython` is as simple as:

* Add `vpython` to `PATH`.
* Write an enviornment specification naming packages.
* Change tool invocation from `python` to `vpython`.

Using `vpython` offers several benefits to standard Python, especially when
vendoring packages. Notably, with `vpython`:

* It is trivially enable hermetic Python everywhere.
* No `sys.path` manipulation is needed to load vendored or imported packages.
* Any tool can trivially define which package(s) it needs, without requiring
  coordination or cooperation from other tools.
* Adding new Python dependencies to a project is non-invasive and immediate.
* Package downloading and deployment are baked into `vpython` and built on
  fast and secure Google Cloud Platform technologies.
* No more custom bootstraps. Several projects and tools, including multiple
  within the infra code base, have bootstrap scripts that vendor packages or
  mimic a VirtualEnv. These are at best repetitive and, at worst, buggy and
  insecure.
* Depenencies are explicitly stated, not assumed.

### Why VirtualEnv?

VirtualEnv offers several benefits over system Python. Primarily, it is the
community-supported mechanism to vendor Python packages. This means that we're
using packages the way they are written and tested to be used, no hacks needed.

By using the same environemnt everywhere, Python invocations become
reproducible. A tool run on a developer's system will load the same versions
of the same libraries as it will on a production system. A production system
will no longer fail because it is missing a package, or because it has the
wrong version.

The classic mechanism for vendoring, `sys.path` manipulation, is nuanced, buggy,
and unsupported by the Python community. It is difficult to get right on all
platforms in all environments for all packages. A notorious example of this is
`protobuf` and other domain-bound packages, which actively fight `sys.path`
inclusion. Using VirtualEnv means that any compliant Python package can
trivially be included into a project.

### Why CIPD?

[CIPD](https://github.com/luci/luci-go/tree/master/cipd) is Chrome's
infrastructure package deployment system. It is simple, accessible, fast, and
backed by resilient systems such as Google Storage and AppEngine.

Unlike `pip`, a CIPD package is defined by its content, enabling precise package
matching instead of fuzzy version matching (e.g., `numpy >= 1.2`, and
`numpy == 1.2` both can match multiple `numpy` packages in `pip`).

CIPD also supports ACLs, enabling privileged Python projects to easily vendor
sensitive packages.

### Why wheels?

A Python [wheel](https://www.python.org/dev/peps/pep-0427/) is a simple binary
distrubition of Python code. A wheel can be generic (pure Python) or system-
and architecture-bound (e.g., 64-bit Mac OSX).

Wheels are prefered over eggs because they come packaged with compiled binaries.
This makes their deployment simple (unpack via `pip`) and reduces system
requirements and variation, since local compilation is not needed.

The increased management burden of maintaining separate wheels for the same
package, one for each architecture, is handled naturally by CIPD, removing the
only real pain point.

## Wheel Guidance

This section contains recommendations for building or uploading wheel CIPD
packages, including platform-specific guidance.

CIPD wheel packages are CIPD packages that contain Python wheels. A given CIPD
package can contain multiple wheels for multiple platforms, but should only
contain one version of any given package for any given architecture/platform.

For example, you can bundle a Windows, Linux, and Mac OSX version of `numpy` and
`coverage` in the same CIPD package, but you should not bundle `numpy==1.11` and
`numpy==1.12` in the same package.

The reason for this is that `vpython` identifies which wheels to install by
scanning the contents of the CIPD package, and if multiple versions appear,
there is no clear guidance about which should be used.

### Mac OSX

Use the `m` ABI suffix and the `macosx_...` platform. `vpython` installs wheels
with the `--force` flag, so slight binary incompatibilities (e.g., specific OSX
versions) can be glossed over.

    coverage-4.3.4-cp27-cp27m-macosx_10_10_x86_64.whl

### Linux

Use wheels with the `mu` ABI suffix and the `manylinux1` platform. For example:

    coverage-4.3.4-cp27-cp27mu-manylinux1_x86_64.whl

### Windows



## Setup and Invocation

`vpython` can be invoked by replacing `python` in the command-line with
`vpython`.

`vpython` works with a default Python environment out of the box. To add
vendored packges, you need to define an enviornment specification file that
describes which wheels to install.

An enviornment specification file is a text protobuf defined as `Spec`
[here](./api/env/spec.proto). An example is:

```
# Any 2.7 interpreter will do.
python_version: "2.7"

# Include "numpy" for the current architecture.
wheel {
  path: "infra/python/wheels/numpy/${platform}-${arch}"
  version: "version:1.11.0"
}

# Include "coverage" for the current architecture.
wheel {
  path: "infra/python/wheels/coverage/${platform}-${arch}"
  version: "version:4.1"
}
```

This specification can be supplied in one of three ways:

* Explicitly, as a command-line option to `vpython`.
* Implicitly, as a file alongside your entry point. For example, if you are
  running `test_runner.py`, `vpython` will look for `test_runner.py.vpython`
  next to it and load the environment from there.
* Implicitly, inined in your main file. `vpython` will scan the main entry point
  for sentinel text and, if present, load the specification from that.

### Migration

#### Command-line.

`vpython` is a natural replacement for `python` in the command line:

```sh
python ./foo/bar/baz.py -d --flag value arg arg whatever
```

Becomes:
```sh
vpython ./foo/bar/baz.py -d --flag value arg arg whatever
```

The `vpython` tool accepts its own command-line arguments. In this case, use
a `--` seprator to differentiate between `vpython` options and `python` options:

```sh
vpython -spec /path/to/spec.vpython -- ./foo/bar/baz.py
```

#### Shebang (POSIX)

If your script uses implicit specification (file or inline), replacing `python`
with `vpython` in your shebang line will automatically work.

```sh
#!/usr/bin/env vpython
```

