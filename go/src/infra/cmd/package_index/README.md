package_index generates kzips for [Kythe](https://kythe.io/) to use in
generating cross-references in Code Search.

Googlers: For more information on package_index and its usage, refer to the
[documentation](https://g3doc.corp.google.com/company/teams/chrome/ops/core_infra/source/codesearch/running_package_index_locally.md?cl=head)
on the ChOps Eng portal.

# Developing

To build and run package_index,
```shell
$ go build
$ ./package_index \
  --checkout_dir ~/chromium/src \
  --out_dir src/out/Debug \
  --path_to_compdb ~/chromium/src/out/Debug/compile_commands.json \
  --path_to_gn_targets ~/chromium/src/out/Debug/gn_targets.json \
  --keep_filepaths_files
  --corpus chromium.googlesource.com/chromium/src \
  --build_config linux \
  --path_to_archive_output ~/test.kzip \
```

You can test package_index by running
```shell
$ go test -run TestPackageIndexUnix # or TestPackageIndexWindows
```

This script is invoked as part of the gen builder, right towards the end. It
takes as input various artifacts produced during earlier parts of the build.

Let's walk through this line-by-line, to explain what each part is for, and
what we must do for the necessary input to exist. Note that here we assume you
have the build repo checked out under `~/infra/build`.

*   `--checkout_dir ~/chromium/src`

This points to the root of your Chromium checkout. Again, we assume a
particular location, modify as necessary. Any commands listed later on should
be run from this directory, unless otherwise specified.

* `--out_dir src/out/Debug`

We need to build chromium before running this script, as we package up
generated files into our kzip. Any build config will work, but `Debug` is
used for the production pipeline, so we use it here too.

If you're interested in Mojom, you need to add
`enable_kythe_annotations = true` to your `args.gn`, and then run
`$ infra/build/scripts/slave/recipe_modules/codesearch/resources/add_kythe_metadata.py ~/chromium/src/out/Debug/gen`
for the relevant annotations to be present. If not, you can skip this step.

*   `--path_to_compdb ~/chromium/src/out/Debug/compile_commands.json`

We read the compilation database to determine which files are built and with
what args. You can generate this file by running the following:

`$ ninja -C out/Debug chromium -t compdb cc cxx objc objcxx > out/Debug/compile_commands.json`

*   `--path_to_gn_targets ~/chromium/src/out/Debug/gn_targets.json`

We also read a dump of the gn targets to find all the Mojom targets, as their
build rules aren't as straightforward. This can be generated by running the
following:

`$ gn desc out/Debug '*' --format=json > out/Debug/gn_targets.json`

*   `--keep_filepaths_files`

We use "filepaths" files to find all the transitive includes for each file. By
default they're deleted each time, and this parameter disables that, which is
useful if you want to run the script multiple times without recreating these
files each time. To generate these files, you first must download the
`translation_unit` clang tool:

`$ ~/chromium/src/tools/clang/scripts/update.py --package=translation_unit`

Then you run this tool over all the files of interest via:

`$ tools/clang/scripts/run_tool.py --tool translation_unit -p out/Debug`

Note that this reads from the compilation database, so you must generate this as
described above before running this tool.

*   `--corpus chromium.googlesource.com/chromium/src`
*   `--build_config linux`

These are parameters written into fields of the kzip units. Their values are
arbitrary strings, the values given here match the production configuration.

*   `--path_to_archive-output ~/test.kzip`

Finally, we specify where the output should be written.

# Adding a new language
If you're adding a new language, you'll likely want to create a
language-specific target struct that's used to implement `GnTargetInterface`.
The struct should contain all information needed to implement the interface,
which defines `getUnit()` that returns a generated CompilationUnit for a
target, and `getFiles()` that returns a list of all files required for
compilation of a target.

Once you're done, populate package_index_testdata which contains the following
directories:
  *  `input/src`: Contains inputs to tests.
  *  `files.expected`: Expected data files which are the output of running
  package_index with `input/src`. As mentioned in
  https://kythe.io/docs/kythe-kzip.html, the name of each file is named its
  SHA256 hash.
  *  `units.expected`: Expected compilation units which are the output of
  running package_index with `input/src`.
  *  `units_win.expected`: Same as units.expected, but for Windows instead of
  Linux.

You may find https://crrev.com/c/3288987, which adds torque to package_index,
useful as reference.

# Uploading to CIPD

Release is handled automatically but if you'd like to manually upload to CIPD,
the linux and windows binaries are located under
[infra/tools/package_index](https://chrome-infra-packages.appspot.com/p/infra/tools/package_index/).
To upload an updated version to CIPD, first navigate to
[infra/infra/build/packages](https://source.chromium.org/chromium/infra/infra/+/master:build/packages/).

If building for linux, run
```shell
GOOS=linux GOARCH=amd64  vpython ../build.py --upload package_index.yaml
```

Or, if building for windows, run
```shell
GOOS=windows GOARCH=amd64  vpython ../build.py --upload package_index.yaml
```

You may need to check CIPD to make sure the instances uploaded are set to the
latest ref.