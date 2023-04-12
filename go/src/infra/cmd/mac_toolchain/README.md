# Ephemeral Xcode Installer for Mac

The purpose of this binary is to provide tools for temporary Xcode
installations, e.g. to build and test Chrome / Chromium on Mac OS X and iOS.

Specifically, it provides tools for creating appropriate Xcode CIPD packages and
then installing them on demand on bots and developer machines.

## Installation

### From `depot_tools`

The `mac_toolchain` tool is available in
[`depot_tools`](https://chromium.googlesource.com/chromium/tools/depot_tools.git)
as a shim which automatically installs the binary on Mac OS X.

### Prebuilt CIPD Package

For `depot_tool`-independent installation (e.g. on a bot, in a recipe), the
preferred method of installing this tool is through CIPD, e.g.:

    $ echo 'infra/tools/mac_toolchain/${platform}  latest' | cipd ensure -ensure-file - -root .

will install the `mac_toolchain` binary in the current directory (as specified
by the `-root` argument).

Note, that the CIPD package currently exists only for Mac OS, and not for any
other platform, since Xcode installation only makes sense on a Mac.

You can also create shim scripts to install the actual binaries automatically,
and optionally pin the `mac_toolchain` package to a specific revision. Pinning
to `latest` will always install the latest available revision. For inspiration,
see
[`cipd_manifest.txt`](https://chromium.googlesource.com/chromium/tools/depot_tools.git/+/master/cipd_manifest.txt)
in `depot_tools`, and the corresponding shims,
e.g. [`vpython`](https://chromium.googlesource.com/chromium/tools/depot_tools.git/+/master/vpython).

The prebuilt CIPD package is configured and built automatically for (almost)
every committed revision of `infra.git`. It's configuration file is
[`mac_toolchain.yaml`](https://chromium.googlesource.com/infra/infra/+/master/build/packages/mac_toolchain.yaml).

### Compile from source

Since this is a standard Go package, you can also install it by running `go
install` in this folder. See the [`infra/go/README.md`](../../../../README.md) file
for the Go setup.

## Working with Xcode packages

### Installing an Xcode package

    mac_toolchain install -xcode-version XXXX -output-dir /path/to/root

Add `-kind mac` or `-kind ios` argument to install a package for mac or ios
tasks. ~~iOS kind has iOS SDK and default iOS runtime installed additionally.~~
(**As of 2023, Xcode is now uploaded as a whole to the "mac" package if you are on MacOS13+, so there is
no difference between mac and ios kind -- both will install the entire Xcode.app
with iOS runtime included**) If
not specified, the tool installs a mac kind.

This will install the requested version of `Xcode.app` in the `/path/to/root`
folder.  Run `mac_toolchain help install` for more information.

_Note:_ to access the Xcode packages, you may need to run:

    cipd auth-login

or pass the appropriate `-service-account-json` argument to `cipd ensure`.

### Creating a new Xcode package

Download the Xcode zip file from your Apple's Developer account, unpack it, and
point the script at the resulting `Xcode.app`.

    mac_toolchain upload -xcode-path /path/to/Xcode.app

This will split up the `Xcode.app` container into several CIPD packages and will
upload and tag them properly. Run `mac_toolchain help upload` for more options.

The upload command is meant to be run manually, and it will upload many GB of
data. Be patient.

~~By changing `CFBundleShortVersionString` and `ProductBuildVersion` in
`Xcode.app/Contents/version.plist`, you can customize the Xcode version in CIPD
`ref`s and `tag`s attached to Xcode when uploading.~~<br>
**Update**: changing the contents in Xcode.app is no longer recommended due to the
new codesign check added in macOS 13. If the Xcode will be installed and run on
macOS 13+, please don't change anything in `Xcode.app/Contents/version.plist`.

By passing in `-skip-ref-tag` argument, the uploaded packages will have no CIPD
`ref` or `tag` attached.

## Working with iOS runtime packages

mac_toolchain uploads the default iOS runtime and the rest of Xcode to different
CIPD packages. It's also able to download an Xcode package without any iOS
runtime bundled.

This is useful when we want to assemble an Xcode with it's core program and the
only runtime needed for iOS test tasks in infra.

Runtime related features are added at around mid 2021. Design and related
changes were tracked in crbug.com/1191260.

### Runtime types

Two types of iOS runtimes are stored in CIPD path
`infra_internal/ios/xcode/ios_runtime`, distinguished through different
combinations of tags and refs:

- `manually_uploaded` type: Uploaded with `mac_toolchain upload-runtime`
 command.

- `xcode_default` type: Uploaded when uploading an Xcode. Packages of this type
are labeled with both runtime version & Xcode version.

### Installing a runtime

Use `mac_toolchain install-runtime`. Either or both runtime version and Xcode
version are accepted as input. Please read `mac_toolchain help install-runtime`
to learn the logic of choosing an appropriate runtime with given input
combinations.

### Creating a runtime package

When uploading an Xcode, the runtime bundled in the Xcode version is uploaded
automatically as an `xcode_default` package.

Use `mac_toolchain upload-runtime` to upload a `manually_uploaded` runtime
package. It's recommended to only use this command on the official runtimes
downloaded through Xcode to system library
`/Library/Developer/CoreSimulator/Profiles/Runtimes/`.

### Installing an Xcode package without runtime

~~Use following command to download an Xcode package without runtime~~

```
mac_toolchain install -kind ios -xcode-version SOME_VERSION -output-dir path/to/Xcode.app -with-runtime=False
```

~~This is specifically used in iOS tester machines across chromium infra where
mac_toolchain is invoked to download Xcode. iOS test runner further downloads
the requested runtime in task and assembles a full Xcode package with runtime
from multiple packages downloaded (or read from cache).~~<br>
**Update (2023)**: in MacOS13+, all Xcode.app contents will be bundled in the "mac" package,
so the above command will still install the runtime to the output directory.



## Debugging packages

To debug the packages locally (for e.g. uploading an Xcode), run:

    mac_toolchain package -output-dir path/to/dir -xcode-path /path/to/Xcode.app

This will drop `mac.cipd`, `ios.cipd`, `ios_runtime.cipd` files in `path/to/out`
directory and will not try to upload the packages to CIPD server.

You can then install Xcode from these local packages with:

    cipd pkg-deploy -root path/to/Xcode.app path/to/out/mac.cipd
    cipd pkg-deploy -root path/to/Xcode.app path/to/out/ios.cipd
    cipd pkg-deploy -root path/to/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes path/to/out/ios_runtime.cipd
