# vpython3

This is a reimplementation of vpython targeting drop-in replacement. We
tried to keep most of the features same, with some exceptions:
- Environment VPYTHON_CLEAR_PYTHONPATH is not supported.
- Argument -vpython-interpreter is not supported.
- Argument -vpython-tool is partially supported.
  - Subcommand install: Argument -name is not supported.
  - Subcommand verify: No longer has a default set of verified tags.
  - Subcommand delete: Not supported.
- Spec config python_version is not supported.
- Spec config virtualenv is not supported.
