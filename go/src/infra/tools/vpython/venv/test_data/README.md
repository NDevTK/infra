## Test Data?

This test data directory is used by VirtualEnv tests to simulate a full
environment setup.

A fake CIPD client in `venv_test.go` will load this, treating each intermediate
file as a CIPD archive:

* If the file is a directory, it will be recursively copied when installed.
* If the file is a ZIP archive (ends with `.zip`), it will be unzipped when
  installed.

It sucks that we're committing some binaries to the repository, but it's the
best option to have a cross-platform self-contained unit test that really
exercises the interesting parts of the setup. The `virtualenv` binary is the
only large binary, though. It's downloaded directly from the Internet.

The other binaries are wheels, generated from source that is also checked into
the `test_data` directory. These wheels have equivalent `.src` components that
were used to generate them.

To build a wheel from source, `cd` into a source directory and run:

    $ python setup.py bdist_wheel

The wheel will be created in `/dist/`.

However, note that it is unlikely any of the wheels will actually need to be
regenerated, since they are simple artifacts.
