# Adding a new wheel for vpython

Adding a new vpython wheel is often just a single line of code, but
in some cases may be more complicated. This doc provides a step-by-step
flow for common and less-common situations.

1. First, go to https://pypi.org/ and find the wheel at the version you will be adding.
   Click "Download files" to see how the wheel is distributed.

1. Determine what type of wheel it is:
   1. Universal (prebuilt)
      * Pure-python libraries already packaged as wheels.
      * These will have a `*-py2.py3-none-any.whl` file (may be just py3)
      * Example: https://pypi.org/project/requests/#files
      * See instructions in [Adding A Universal Wheel](#adding-a-universal-wheel).
   1. Source archives
      * These will have a `*.tar.gz` file or rarely another type of archive.
         You'll have to fetch this tarball and look to see if it contains any
         .c or .cc files. If it does not, then follow the instructions
         in
         [Adding A Universal Source Wheel](#adding-a-universalsource-wheel).
      * If there are .c or .cc files, instead follow the instructions
         in [Adding A SourceOrPrebuilt Wheel](#adding-a-sourceorprebuilt-wheel).
   1. Special wheels
      * Occasionally a wheel will require custom build steps beyond what the
        standard classes provide. An example of this is [mpi4py](./wheel_mpi4py.py).
        By implementing a custom build_fn, arbitrary steps can be executed.
        This should be last resort; please ask for help in a bug before
        implementing a wheel this way.

1. If you are adding a new version of an existing wheel, please leave the old versions
   and add a new entry for the new version. This helps keep [wheels.md](./wheels.md)
   as a catalog of available wheels.
1. Once you are done adding the wheel, run `vpython3 -m infra.tools.dockerbuild wheel-dump` to update `wheels.md` before creating your CL.
1. The CL tryjobs will verify that the wheel builds on all of the platforms.
1. Once the CL is committed, the production builders will build and upload the wheel to CIPD.

# Adding A Universal Wheel

For Universal wheels, add a line like the following to the appropriate
section of [wheels.py](./wheels.py) (for the `attrs` wheel):

        Universal('attrs', '21.4.0'),

If the wheel file is named with `-py3-none-any` (rather than
`py2.py3-none-any`), include a `pyversions` attribute:

        Universal('cachetools', '4.2.2', pyversions=['py3']),

dockerbuild will simply download and package the prebuilt universal wheel.

In the uncommon case that we need to patch the wheel source, the wheel
must be added as a [UniversalSource](#adding-a-universalsource-wheel)
wheel instead.

# Adding A UniversalSource Wheel

UniversalSource wheels can typically be added to the appropriate
section of [wheels.py](./wheels.py) as follows, using the `httplib2`
wheel as an example:

        UniversalSource('httplib2', '0.13.1'),

In the uncommon case that we need to patch the wheel source, patch names
can be listed after the wheel version:

        UniversalSource(
            'clusterfuzz',
            '2.5.6',
            patches=('no-deps-install',),
        ),

See the footnotes on [Custom Patches](#custom-patches) for more information.

# Adding A SourceOrPrebuilt Wheel

SourceOrPrebuilt is used for wheels with compiled code, which we can either
build from source (preferably), or use a prebuilt wheel from pypi.org.

The simplest example looks like this:

        SourceOrPrebuilt(
            'zstandard', '0.16.0', packaged=(), pyversions=['py3']),
    )

The `packaged` attribute is a list of wheel platforms for which a prebuilt
wheel should be used. Prefer to set it to an empty tuple `()`, to build from source on all
platforms. If a given platform is too difficult to build from source, it can
be specified as follows, using the platform names from
[build_platform.py](./build_platform.py):

            packaged=(
                'windows-x86-py3.8',
                'windows-x86-py3.11',
                'windows-x64-py3.8',
                'windows-x64-py3.11',
            ),

This would use prebuilt wheels for all of the Windows platforms.

`pyversions` should typically be set to ['py3'] for newly-added wheels.
It contains 'py2' for some older wheels so as to keep the naming consistent.

Other useful attributes include:

* `only_plat` specifies that the wheel should only be built on the given
platforms.
* `skip_plat` is the reverse of only_plat: the given platforms will not
have the wheel built.
* `patches` gives a list of patches to be applied, see
[Custom Patches](#custom-patches).
* `patch_version` is a patch version to be appended to the wheel version.
It should be incremented whenever patches are changed, or if the build
environment is changed and we want to trigger the wheel to rebuild.
* `tpp_libs` (3pp-libs) is a list of prebuilt library packages to install
from CIPD into the wheel build environment. For example:

            tpp_libs=[('infra/3pp/static_libs/re2',
                       'version:2@2022-12-01.chromium.1')],

When cross-compiling, `tpp_libs` packages are installed for the target
platform.

* `tpp_tools` (3pp-tools) is a list of prebuilt build-time tool packages
to install from CIPD into the wheel build environment. For example:

            tpp_tools=[
                ('infra/3pp/tools/cmake', 'version:2@3.26.0.chromium.7'),
            ],

When cross-compiling, `tpp_tools` packages are installed for the host
platform.

* `arch_map` can be used if the prebuilt wheels do not precisely match
our usual wheel ABI. For example:

            arch_map={'mac-x64-py3.8': ['macosx_10_14_x86_64']},

allows the prebuilt wheel to use the macOS 10.14 ABI, whereas we typically
target 10.13. You'll know that you need to use this if the build fails
because pip is unable to find the prebuilt wheel. Be cautious when overriding
the ABI to a newer OS version, as it means the packaged wheel may not work on
all of the machines in our fleet.

* TODO: describe build_deps

## Custom patches

While we strongly prefer to not patch anything, sometimes we need a backport
or local fix for our system.

Here's the quick overview:

* Patches are only supported with `UniversalSource` and `SourceOrPrebuilt` since
  we need to unpack the source & patch it directly before building the wheel.
* All patches live under `patches/`.
* All patches must be in the `-p1` format.
* The filenames must start with the respective package name & version and end
  in `.patch`.  e.g. `UniversalSource('scandir', '1.9.0')` will have a prefix
  of `scandir-1.9.0-` and a suffix of `.patch`.
* Add the shortnames into the `patches=(...)` tuple to `UniversalSource` or
  `SourceOrPrebuilt`.
* All patches should be well documented in the file header itself.

A short example:

```python
  UniversalSource('scandir', '1.9.0', patches=(
      'some-fix',
      'another-change',
  )),
```

This will apply the two patches:
* `patches/scandir-1.9.0-some-fix.patch`
* `patches/scandir-1.9.0-another-change.patch`
